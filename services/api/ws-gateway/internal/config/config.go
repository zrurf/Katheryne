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

	RedisAddr string
	MQAddr    string

	WSHeartbeatInterval int64
	WSReadTimeout       int64
	WSWriteTimeout      int64
	WSMaxMessageSize    int64
}