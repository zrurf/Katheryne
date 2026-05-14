package svc

import (
	"bot/internal/config"
	"bot/internal/middleware"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config            config.Config
	AuthMiddleware    rest.Middleware
	BotAuthMiddleware rest.Middleware
	Redis             *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	return &ServiceContext{
		Config:            c,
		AuthMiddleware:    middleware.NewAuthMiddleware().Handle,
		BotAuthMiddleware: middleware.NewBotAuthMiddleware().Handle,
		Redis:             redisClient,
	}
}