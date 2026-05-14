package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	RedisAddr string    `json:"RedisAddr"`
	MQAddr    string    `json:"MQAddr"`
	UserDB    DBConfig  `json:"UserDB"`
	Opaque    OpaqueConfig `json:"Opaque"`

	TFATokenExpireSeconds     int `json:"TFATokenExpireSeconds"`
	AccessTokenExpireSeconds  int `json:"AccessTokenExpireSeconds"`
	RefreshTokenExpireSeconds int `json:"RefreshTokenExpireSeconds"`
	MaxTokenRetries           int `json:"MaxTokenRetries"`
	SessionTokenLength        int `json:"SessionTokenLength"`
	MaxUidRetries             int `json:"MaxUidRetries"`

	Snoyflake SonyflakeConfig `json:"Snoyflake"`
}

type DBConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	DBName   string `json:"DBName"`
}

type OpaqueConfig struct {
	OPRFSeedFile        string `json:"OPRFSeedFile"`
	ServerPublicKeyFile string `json:"ServerPublicKeyFile"`
	ServerSecretKeyFile string `json:"ServerSecretKeyFile"`
}

type SonyflakeConfig struct {
	StartTime string `json:"StartTime"`
	MachineID uint16 `json:"MachineID"`
}
