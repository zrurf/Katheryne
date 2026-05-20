package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	RedisAddr       string   `json:"RedisAddr"`
	BotDB           DBConfig `json:"BotDB"`
	PiiSalt         string   `json:"PiiSalt"`
	MaxRetries      int      `json:"MaxRetries"`
	RetryBackoffMin int      `json:"RetryBackoffMin"`
	RetryBackoffMax int      `json:"RetryBackoffMax"`
	DefaultExpiry   int64    `json:"DefaultExpiry"`
}

type DBConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	DBName   string `json:"DBName"`
}
