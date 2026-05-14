package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	RedisAddr string          `json:"RedisAddr"`
	MQAddr    string          `json:"MQAddr"`
	MessageDB DBConfig        `json:"MessageDB"`
	UserRpc   zrpc.RpcClientConf `json:"UserRpc"`
}

type DBConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	DBName   string `json:"DBName"`
}
