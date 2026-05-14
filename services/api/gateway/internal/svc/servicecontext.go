package svc

import (
	"auth/authclient"
	"conversation/conversationclient"
	"gateway/internal/config"
	"gateway/internal/middleware"
	"message/messageclient"
	"oss/ossclient"
	"social/socialclient"
	"user/userclient"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware
	Redis          *redis.Client

	AuthRpc         authclient.Auth
	ConversationRpc conversationclient.Conversation
	MessageRpc      messageclient.Message
	OssRpc          ossclient.OSS
	SocialRpc       socialclient.Social
	UserRpc         userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	return &ServiceContext{
		Config:          c,
		AuthMiddleware:  middleware.NewAuthMiddleware(redisClient).Handle,
		Redis:           redisClient,
		AuthRpc:         authclient.NewAuth(zrpc.MustNewClient(c.AuthRpc)),
		ConversationRpc: conversationclient.NewConversation(zrpc.MustNewClient(c.ConversationRpc)),
		MessageRpc:      messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		OssRpc:          ossclient.NewOSS(zrpc.MustNewClient(c.OssRpc)),
		SocialRpc:       socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		UserRpc:         userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}