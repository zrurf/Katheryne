package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	RedisAddr       string             `json:"RedisAddr"`
	BotDB           DBConfig           `json:"BotDB"`
	SocialDB        DBConfig           `json:"SocialDB"`
	ConversationRpc zrpc.RpcClientConf `json:"ConversationRpc"`
	PiiSalt         string             `json:"PiiSalt"`
	MaxRetries      int                `json:"MaxRetries"`
	RetryBackoffMin int                `json:"RetryBackoffMin"`
	RetryBackoffMax int                `json:"RetryBackoffMax"`
	DefaultExpiry   int64              `json:"DefaultExpiry"`
}

type DBConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	DBName   string `json:"DBName"`
}
