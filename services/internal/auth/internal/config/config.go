package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	RedisAddr string
	MQAddr    string
	UserDB    DBConfig
	Opaque    OpaqueConfig

	TFATokenExpireSeconds     int
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int
	MaxTokenRetries           int
	SessionTokenLength        int
	MaxUidRetries             int

	Snoyflake SonyflakeConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type OpaqueConfig struct {
	OPRFSeedFile        string
	ServerPublicKeyFile string
	ServerSecretKeyFile string
}

type SonyflakeConfig struct {
	StartTime string
	MachineID uint16
}
