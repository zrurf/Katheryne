package svc

import (
	"auth/authclient"
	"bot/botclient"
	"conversation/conversationclient"
	"message/messageclient"
	"social/socialclient"
	"user/userclient"
	"ws-gateway/internal/config"
	"ws-gateway/internal/logic/ws"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client

	AuthRpc         authclient.Auth
	UserRpc         userclient.User
	SocialRpc       socialclient.Social
	MessageRpc      messageclient.Message
	ConversationRpc conversationclient.Conversation
	BotRpc          botclient.Bot

	Hub *ws.Hub
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	hub := ws.NewHub(ws.HubConfig{
		Redis:               redisClient,
		HeartbeatInterval:   c.WSHeartbeatInterval,
		ReadTimeout:         c.WSReadTimeout,
		WriteTimeout:        c.WSWriteTimeout,
		MaxMessageSize:      c.WSMaxMessageSize,
		AuthRpc:             authclient.NewAuth(zrpc.MustNewClient(c.AuthRpc)),
		UserRpc:             userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		SocialRpc:           socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		MessageRpc:          messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		ConversationRpc:     conversationclient.NewConversation(zrpc.MustNewClient(c.ConversationRpc)),
		BotRpc:              botclient.NewBot(zrpc.MustNewClient(c.BotRpc)),
	})
	go hub.Run()

	return &ServiceContext{
		Config:          c,
		Redis:           redisClient,
		AuthRpc:         authclient.NewAuth(zrpc.MustNewClient(c.AuthRpc)),
		UserRpc:         userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		SocialRpc:       socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		MessageRpc:      messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		ConversationRpc: conversationclient.NewConversation(zrpc.MustNewClient(c.ConversationRpc)),
		BotRpc:          botclient.NewBot(zrpc.MustNewClient(c.BotRpc)),
		Hub:             hub,
	}
}