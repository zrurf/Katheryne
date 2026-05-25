package svc

import (
	"auth/authclient"
	"bot/botclient"
	"conversation/conversationclient"
	"gateway/internal/config"
	"gateway/internal/middleware"
	"message/messageclient"
	"oss/ossclient"
	"rag/ragclient"
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
	BotRpc          botclient.Bot
	ConversationRpc conversationclient.Conversation
	MessageRpc      messageclient.Message
	OssRpc          ossclient.OSS
	RagRpc          ragclient.Rag
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
		BotRpc:          botclient.NewBot(zrpc.MustNewClient(c.BotRpc)),
		ConversationRpc: conversationclient.NewConversation(zrpc.MustNewClient(c.ConversationRpc)),
		MessageRpc:      messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		OssRpc:          ossclient.NewOSS(zrpc.MustNewClient(c.OssRpc)),
		RagRpc:          ragclient.NewRag(zrpc.MustNewClient(c.RagRpc)),
		SocialRpc:       socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		UserRpc:         userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
