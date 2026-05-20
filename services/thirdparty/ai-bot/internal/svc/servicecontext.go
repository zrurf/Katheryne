package svc

import (
	"ai-bot/internal/bot"
	"ai-bot/internal/config"
	"ai-bot/internal/logic"
	"net/http"
)

type ServiceContext struct {
	Config         config.Config
	BotClient      *bot.Client
	MsgHandler     *logic.MessageHandler
	HealthHandler  *HealthHandler
	MetricsHandler *MetricsHandler
}

func NewServiceContext(c config.Config) *ServiceContext {
	botClient := bot.NewClient(bot.ClientConfig{
		TokenURL:     c.BotAPITokenURL,
		RefreshURL:   c.BotAPIRefreshURL,
		WSGatewayURL: c.WSGatewayURL,
		ClientID:     c.BotClientID,
		ClientSecret: c.BotClientSecret,
	})

	msgHandler := logic.NewMessageHandler(logic.HandlerConfig{
		LLMProvider:    c.LLM.Provider,
		LLMAPIKey:      c.LLM.APIKey,
		LLMBaseURL:     c.LLM.BaseURL,
		LLMModel:       c.LLM.Model,
		LLMMaxTokens:   c.LLM.MaxTokens,
		LLMTemperature: c.LLM.Temperature,
	})

	svcCtx := &ServiceContext{
		Config:     c,
		BotClient:  botClient,
		MsgHandler: msgHandler,
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
