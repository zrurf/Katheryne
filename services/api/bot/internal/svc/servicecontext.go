// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"bot/internal/config"
	"bot/internal/middleware"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config            config.Config
	AuthMiddleware    rest.Middleware
	BotAuthMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:            c,
		AuthMiddleware:    middleware.NewAuthMiddleware().Handle,
		BotAuthMiddleware: middleware.NewBotAuthMiddleware().Handle,
	}
}
