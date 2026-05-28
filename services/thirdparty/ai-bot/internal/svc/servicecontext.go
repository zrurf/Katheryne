package svc

import (
	"ai-bot/internal/config"
	"ai-bot/internal/logic"
	"ai-bot/internal/orchestrator"
	"net/http"

	"bot/botclient"
	"mem/memclient"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	Orchestrator   *orchestrator.Orchestrator
	MsgHandler     *logic.MessageHandler // Default handler for utility APIs (summarize, translate, metrics)
	HealthHandler  *HealthHandler
	MetricsHandler *MetricsHandler
}

func NewServiceContext(c config.Config) *ServiceContext {
	var ragClient ragclient.Rag
	if c.RagRpc.Target != "" || len(c.RagRpc.Endpoints) > 0 || len(c.RagRpc.Etcd.Hosts) > 0 {
		client, err := zrpc.NewClient(c.RagRpc)
		if err != nil {
			logx.Errorf("create rag rpc client failed: %v", err)
		} else {
			ragClient = ragclient.NewRag(client)
		}
	}

	var botRpcClient botclient.Bot
	if c.BotRpc.Target != "" || len(c.BotRpc.Endpoints) > 0 || len(c.BotRpc.Etcd.Hosts) > 0 {
		client, err := zrpc.NewClient(c.BotRpc)
		if err != nil {
			logx.Errorf("create bot rpc client failed: %v", err)
		} else {
			botRpcClient = botclient.NewBot(client)
		}
	}

	var memClient memclient.Mem
	if c.MemRpc.Target != "" || len(c.MemRpc.Endpoints) > 0 || len(c.MemRpc.Etcd.Hosts) > 0 {
		client, err := zrpc.NewClient(c.MemRpc)
		if err != nil {
			logx.Errorf("create mem rpc client failed: %v", err)
		} else {
			memClient = memclient.NewMem(client)
			logx.Infof("mem rpc client connected")
		}
	}

	// Default handler for utility APIs (not tied to any specific bot instance)
	msgHandler := logic.NewMessageHandler(logic.HandlerConfig{
		LLMProvider:    c.LLM.Provider,
		LLMAPIKey:      c.LLM.APIKey,
		LLMBaseURL:     c.LLM.BaseURL,
		LLMModel:       c.LLM.Model,
		LLMMaxTokens:   c.LLM.MaxTokens,
		LLMTemperature: c.LLM.Temperature,
		RagClient:      ragClient,
		MemClient:      memClient,
	})
	msgHandler.SetBotName("Katheryne AI")

	orch := orchestrator.NewOrchestrator(orchestrator.OrchestratorConfig{
		TokenURL:     c.BotAPITokenURL,
		RefreshURL:   c.BotAPIRefreshURL,
		WSGatewayURL: c.WSGatewayURL,
		ClientID:     c.BotClientID,
		ClientSecret: c.BotClientSecret,
		RagClient:    ragClient,
		BotRpcClient: botRpcClient,
		MemClient:    memClient,
		DefaultLLM: orchestrator.LLMDefaults{
			Provider:    c.LLM.Provider,
			BaseURL:     c.LLM.BaseURL,
			APIKey:      c.LLM.APIKey,
			Model:       c.LLM.Model,
			MaxTokens:   c.LLM.MaxTokens,
			Temperature: c.LLM.Temperature,
		},
	})

	svcCtx := &ServiceContext{
		Config:       c,
		Orchestrator: orch,
		MsgHandler:   msgHandler,
	}

	svcCtx.HealthHandler = &HealthHandler{svcCtx: svcCtx}
	svcCtx.MetricsHandler = &MetricsHandler{svcCtx: svcCtx}

	return svcCtx
}

type HealthHandler struct {
	svcCtx *ServiceContext
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"katheryne-ai-bot"}`))
}

type MetricsHandler struct {
	svcCtx *ServiceContext
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"bot_status":"connected","cached_conversations":0}`))
}
