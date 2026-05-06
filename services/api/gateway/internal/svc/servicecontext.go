// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

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

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware

	AuthRpc         authclient.Auth
	ConversationRpc conversationclient.Conversation
	MessageRpc      messageclient.Message
	OssRpc          ossclient.OSS
	SocialRpc       socialclient.Social
	UserRpc         userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		AuthMiddleware:  middleware.NewAuthMiddleware(c.RedisAddr).Handle,
		AuthRpc:         authclient.NewAuth(zrpc.MustNewClient(c.AuthRpc)),
		ConversationRpc: conversationclient.NewConversation(zrpc.MustNewClient(c.ConversationRpc)),
		MessageRpc:      messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		OssRpc:          ossclient.NewOSS(zrpc.MustNewClient(c.OssRpc)),
		SocialRpc:       socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		UserRpc:         userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
