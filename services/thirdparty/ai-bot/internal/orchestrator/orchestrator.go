package orchestrator

import (
	"context"
	"mem/memclient"
	"sync"
	"sync/atomic"
	"time"

	"ai-bot/internal/bot"
	"ai-bot/internal/logic"
	"bot/botclient"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// LLMDefaults are fallback LLM settings when instance doesn't specify them
type LLMDefaults struct {
	Provider    string
	BaseURL     string
	MaxTokens   int
	Temperature float64
}

// OrchestratorConfig holds configuration for the multi-tenant orchestrator
type OrchestratorConfig struct {
	TokenURL     string
	RefreshURL   string
	WSGatewayURL string
	ClientID     string
	ClientSecret string
	RagClient    ragclient.Rag
	BotRpcClient botclient.Bot
	MemClient    memclient.Mem
	DefaultLLM   LLMDefaults
}

// RuntimeConfig holds a bot instance's configuration fetched from bot service
type RuntimeConfig struct {
	BotID             int64
	InstanceID        int64
	Name              string
	ClientID          string
	ClientSecret      string
	ModelProvider     string
	ModelName         string
	APIKey            string
	APIBaseURL        string
	SystemPrompt      string
	ConversationStyle string
	ToolDefinitions   string
	KBConfig          string
	KBIDs             []string
	IsOfficial        bool
}

// managedInstance represents a running bot instance managed by the orchestrator
type managedInstance struct {
	config     *RuntimeConfig
	wsClient   *bot.Client
	msgHandler *logic.MessageHandler
	cancel     context.CancelFunc
}

// Orchestrator manages multiple hosted bot instances in a single ai-bot process.
type Orchestrator struct {
	config    OrchestratorConfig
	mu        sync.RWMutex
	instances map[int64]*managedInstance // botID → instance
	running   atomic.Bool
}

func NewOrchestrator(cfg OrchestratorConfig) *Orchestrator {
	return &Orchestrator{
		config:    cfg,
		instances: make(map[int64]*managedInstance),
	}
}

// Start loads all hosted instances from the bot service and starts them.
func (o *Orchestrator) Start() error {
	if o.config.BotRpcClient == nil {
		logx.Info("orchestrator: BotRpcClient not configured, skipping multi-tenant mode")
		return nil
	}

	o.running.Store(true)

	// Load all hosted instances
	if err := o.syncInstances(); err != nil {
		logx.Errorf("orchestrator: initial sync failed: %v", err)
		return err
	}

	// Start background watcher (poll every 60s)
	go o.watchLoop()

	logx.Infof("orchestrator: started, managing %d instances", o.count())
	return nil
}

// Stop shuts down all managed instances.
func (o *Orchestrator) Stop() {
	o.running.Store(false)
	o.mu.Lock()
	defer o.mu.Unlock()

	for botID, inst := range o.instances {
		logx.Infof("orchestrator: stopping instance bot_id=%d", botID)
		if inst.cancel != nil {
			inst.cancel()
		}
		inst.wsClient.Stop()
	}
	o.instances = make(map[int64]*managedInstance)
	logx.Info("orchestrator: all instances stopped")
}

func (o *Orchestrator) count() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.instances)
}

// syncInstances fetches hosted instances from bot service and starts/stops as needed
func (o *Orchestrator) syncInstances() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := o.config.BotRpcClient.ListHostedInstances(ctx, &botclient.ListHostedInstancesReq{})
	if err != nil {
		return err
	}

	desired := make(map[int64]bool)
	for _, inst := range resp.List {
		desired[inst.BotId] = true

		o.mu.RLock()
		_, exists := o.instances[inst.BotId]
		o.mu.RUnlock()

		if !exists {
			if err := o.startInstance(inst.BotId); err != nil {
				logx.Errorf("orchestrator: failed to start instance bot_id=%d: %v", inst.BotId, err)
			}
		}
	}

	// Stop instances that no longer exist
	o.mu.Lock()
	for botID := range o.instances {
		if !desired[botID] {
			inst := o.instances[botID]
			logx.Infof("orchestrator: removing instance bot_id=%d", botID)
			if inst.cancel != nil {
				inst.cancel()
			}
			inst.wsClient.Stop()
			delete(o.instances, botID)
		}
	}
	o.mu.Unlock()

	return nil
}

// startInstance fetches runtime config and starts a WS connection for one bot instance
func (o *Orchestrator) startInstance(botID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	runtimeCfg, err := o.config.BotRpcClient.GetBotRuntimeConfig(ctx, &botclient.GetBotRuntimeConfigReq{
		BotId: botID,
	})
	if err != nil {
		return err
	}

	cfg := &RuntimeConfig{
		BotID:             botID,
		InstanceID:        runtimeCfg.InstanceId,
		Name:              runtimeCfg.Name,
		ClientID:          runtimeCfg.ClientId,
		ClientSecret:      runtimeCfg.ClientSecret,
		ModelProvider:     runtimeCfg.ModelProvider,
		ModelName:         runtimeCfg.ModelName,
		APIKey:            runtimeCfg.ApiKey,
		APIBaseURL:        runtimeCfg.ApiBaseUrl,
		SystemPrompt:      runtimeCfg.SystemPrompt,
		ConversationStyle: runtimeCfg.ConversationStyle,
		ToolDefinitions:   runtimeCfg.ToolDefinitions,
		KBConfig:          runtimeCfg.KbConfig,
		IsOfficial:        runtimeCfg.IsOfficial,
	}

	if runtimeCfg.KbIds != "" {
		cfg.KBIDs = splitNonEmpty(runtimeCfg.KbIds, ",")
	}

	// Build WS client config using the instance's OAuth credentials
	wsCfg := bot.ClientConfig{
		TokenURL:     o.config.TokenURL,
		RefreshURL:   o.config.RefreshURL,
		WSGatewayURL: o.config.WSGatewayURL,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	}

	wsClient := bot.NewClient(wsCfg)

	// Build MessageHandler with instance-specific config
	handlerCfg := logic.HandlerConfig{
		RagClient: o.config.RagClient,
		MemClient: o.config.MemClient,
		KBIDs:     cfg.KBIDs,
	}
	handlerCfg.LLMProvider = cfg.ModelProvider
	if handlerCfg.LLMProvider == "" {
		handlerCfg.LLMProvider = o.config.DefaultLLM.Provider
	}
	handlerCfg.LLMBaseURL = cfg.APIBaseURL
	if handlerCfg.LLMBaseURL == "" {
		handlerCfg.LLMBaseURL = o.config.DefaultLLM.BaseURL
	}
	handlerCfg.LLMAPIKey = cfg.APIKey
	handlerCfg.LLMModel = cfg.ModelName
	handlerCfg.LLMMaxTokens = o.config.DefaultLLM.MaxTokens
	handlerCfg.LLMTemperature = o.config.DefaultLLM.Temperature

	msgHandler := logic.NewMessageHandler(handlerCfg)
	msgHandler.SetBotID(botID)
	msgHandler.SetOfficial(cfg.IsOfficial)
	if cfg.SystemPrompt != "" {
		msgHandler.SetSystemPrompt(cfg.SystemPrompt)
	}

	// Set up the handler chain
	wsClient.SetHandler(msgHandler)
	msgHandler.SetSender(wsClient)

	// Start the WS connection
	if err := wsClient.Start(); err != nil {
		return err
	}

	bgCtx, bgCancel := context.WithCancel(context.Background())
	msgHandler.StartCleanup(bgCtx, 10*time.Minute)

	inst := &managedInstance{
		config:     cfg,
		wsClient:   wsClient,
		msgHandler: msgHandler,
		cancel:     bgCancel,
	}

	o.mu.Lock()
	o.instances[botID] = inst
	o.mu.Unlock()

	logx.Infof("orchestrator: started instance bot_id=%d name=%s official=%v",
		botID, cfg.Name, cfg.IsOfficial)
	return nil
}

// watchLoop periodically polls for new/removed hosted instances
func (o *Orchestrator) watchLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for o.running.Load() {
		select {
		case <-ticker.C:
			if err := o.syncInstances(); err != nil {
				logx.Errorf("orchestrator: sync failed: %v", err)
			}
		}
	}
}

// splitNonEmpty splits a string and filters empty entries
func splitNonEmpty(s, sep string) []string {
	var result []string
	for _, part := range splitFallback(s, sep) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// splitFallback is a simple split (avoid importing strings in case it's not needed)
func splitFallback(s, sep string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
		}
	}
	parts = append(parts, s[start:])
	return parts
}
