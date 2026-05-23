package handler

import (
	"net/http"

	"ai-bot/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	botMgr := NewBotManagerHandler(svcCtx)
	botInteract := NewBotInteractHandler(svcCtx)

	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/health",
			Handler: svcCtx.HealthHandler.ServeHTTP,
		},
		{
			Method:  http.MethodGet,
			Path:    "/metrics",
			Handler: svcCtx.MetricsHandler.ServeHTTP,
		},
		// Bot management
		{
			Method:  http.MethodGet,
			Path:    "/bot/config",
			Handler: botMgr.GetConfigHandler(),
		},
		{
			Method:  http.MethodPut,
			Path:    "/bot/config",
			Handler: botMgr.UpdateConfigHandler(),
		},
		{
			Method:  http.MethodGet,
			Path:    "/bot/stats",
			Handler: botMgr.GetStatsHandler(),
		},
		{
			Method:  http.MethodGet,
			Path:    "/bot/memory",
			Handler: botMgr.GetMemoryHandler(),
		},
		{
			Method:  http.MethodPost,
			Path:    "/bot/memory/clear",
			Handler: botMgr.ClearMemoryHandler(),
		},
		// Bot AI interaction
		{
			Method:  http.MethodPost,
			Path:    "/bot/summarize",
			Handler: botInteract.SummarizeHandler(),
		},
		{
			Method:  http.MethodPost,
			Path:    "/bot/translate",
			Handler: botInteract.TranslateHandler(),
		},
		{
			Method:  http.MethodPost,
			Path:    "/bot/suggest",
			Handler: botInteract.SuggestHandler(),
		},
		{
			Method:  http.MethodPost,
			Path:    "/bot/moderate",
			Handler: botInteract.ModerateHandler(),
		},
	})
}
