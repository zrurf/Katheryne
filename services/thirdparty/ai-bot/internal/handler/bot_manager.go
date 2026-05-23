package handler

import (
	"ai-bot/internal/logic"
	"ai-bot/internal/svc"
	"encoding/json"
	"net/http"
	"time"

	xhttp "github.com/zeromicro/x/http"
)

type BotManager struct {
	svcCtx *svc.ServiceContext
}

func NewBotManagerHandler(svcCtx *svc.ServiceContext) *BotManager {
	return &BotManager{svcCtx: svcCtx}
}

type BotConfigInfo struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	APIKey      string  `json:"api_key,omitempty"`
	BaseURL     string  `json:"base_url"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Status      string  `json:"status"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
}

type UpdateConfigRequest struct {
	Provider    string  `json:"provider,optional"`
	Model       string  `json:"model,optional"`
	APIKey      string  `json:"api_key,optional"`
	BaseURL     string  `json:"base_url,optional"`
	MaxTokens   int     `json:"max_tokens,optional"`
	Temperature float64 `json:"temperature,optional"`
	Name        string  `json:"name,optional"`
	Description string  `json:"description,optional"`
}

var botName = "Katheryne Bot"
var botDescription = "Katheryne 官方内置 AI 助手，支持智能问答、对话总结、实时翻译、内容审核等功能。"
var startTime = time.Now()

func (h *BotManager) GetConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := h.svcCtx.Config
		info := BotConfigInfo{
			Provider:    cfg.LLM.Provider,
			Model:       cfg.LLM.Model,
			BaseURL:     cfg.LLM.BaseURL,
			MaxTokens:   cfg.LLM.MaxTokens,
			Temperature: cfg.LLM.Temperature,
			Status:      h.svcCtx.MsgHandler.GetStatus(),
			Name:        botName,
			Description: botDescription,
		}
		xhttp.JsonBaseResponseCtx(r.Context(), w, info)
	}
}

func (h *BotManager) UpdateConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdateConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		updated := false
		if req.Provider != "" {
			h.svcCtx.Config.LLM.Provider = req.Provider
			updated = true
		}
		if req.Model != "" {
			h.svcCtx.Config.LLM.Model = req.Model
			updated = true
		}
		if req.APIKey != "" {
			h.svcCtx.Config.LLM.APIKey = req.APIKey
			updated = true
		}
		if req.BaseURL != "" {
			h.svcCtx.Config.LLM.BaseURL = req.BaseURL
			updated = true
		}
		if req.MaxTokens > 0 {
			h.svcCtx.Config.LLM.MaxTokens = req.MaxTokens
			updated = true
		}
		if req.Temperature > 0 {
			h.svcCtx.Config.LLM.Temperature = req.Temperature
			updated = true
		}
		if req.Name != "" {
			botName = req.Name
		}
		if req.Description != "" {
			botDescription = req.Description
		}

		if updated {
			h.svcCtx.MsgHandler.ReloadConfig(logic.HandlerConfig{
				LLMProvider:    h.svcCtx.Config.LLM.Provider,
				LLMAPIKey:      h.svcCtx.Config.LLM.APIKey,
				LLMBaseURL:     h.svcCtx.Config.LLM.BaseURL,
				LLMModel:       h.svcCtx.Config.LLM.Model,
				LLMMaxTokens:   h.svcCtx.Config.LLM.MaxTokens,
				LLMTemperature: h.svcCtx.Config.LLM.Temperature,
			})
		}

		xhttp.JsonBaseResponseCtx(r.Context(), w, map[string]interface{}{
			"success": true,
			"message": "配置已更新",
		})
	}
}

func (h *BotManager) GetStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := h.svcCtx.MsgHandler.GetStats()
		stats["uptime"] = time.Since(startTime).String()
		stats["provider"] = h.svcCtx.Config.LLM.Provider
		stats["model"] = h.svcCtx.Config.LLM.Model
		xhttp.JsonBaseResponseCtx(r.Context(), w, stats)
	}
}

func (h *BotManager) GetMemoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		convID := r.URL.Query().Get("conv_id")
		memory := h.svcCtx.MsgHandler.GetMemory(convID)
		xhttp.JsonBaseResponseCtx(r.Context(), w, map[string]interface{}{
			"conversations": memory,
		})
	}
}

func (h *BotManager) ClearMemoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		convID := r.URL.Query().Get("conv_id")
		h.svcCtx.MsgHandler.ClearMemory(convID)
		xhttp.JsonBaseResponseCtx(r.Context(), w, map[string]interface{}{
			"success": true,
			"message": "记忆已清除",
		})
	}
}
