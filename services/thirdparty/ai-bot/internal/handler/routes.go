package handler

import (
	"net/http"

	"ai-bot/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
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
	})
}
