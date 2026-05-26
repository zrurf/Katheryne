package handler

import (
	"net/http"

	"ws-gateway/internal/handler/internal"
	"ws-gateway/internal/handler/ws"
	"ws-gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/ws",
				Handler: ws.ClientWSHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/ws",
				Handler: ws.BotWSHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/v1/bot"),
	)

	// Internal endpoints (not exposed to clients)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/push_message",
				Handler: internal.PushMessageHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/v1/internal"),
	)
}