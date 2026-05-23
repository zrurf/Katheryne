// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	AuthRpc         zrpc.RpcClientConf
	UserRpc         zrpc.RpcClientConf
	SocialRpc       zrpc.RpcClientConf
	MessageRpc      zrpc.RpcClientConf
	ConversationRpc zrpc.RpcClientConf
	OssRpc          zrpc.RpcClientConf
	BotRpc          zrpc.RpcClientConf

	RedisAddr   string
	AiBotUrl    string
	MaxFileSize int64 `json:",default=104857600"`
}
