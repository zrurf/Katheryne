package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	RedisAddr string    `json:"RedisAddr"`
	RagDB     DBConfig  `json:"RagDB"`
	Qdrant    QdrantConfig `json:"Qdrant"`
	HugeGraph HugeGraphConfig `json:"HugeGraph"`
}

type DBConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	DBName   string `json:"DBName"`
}

type QdrantConfig struct {
	Host      string `json:"Host"`
	Port      int    `json:"Port"`
	UseTLS    bool   `json:"UseTLS"`
	APIKey    string `json:"APIKey,optional"`
	VectorDim int    `json:"VectorDim,default=1024"`
}

type HugeGraphConfig struct {
	BaseURL string `json:"BaseURL"` // e.g. http://hugegraph:8080
	Graph   string `json:"Graph"`   // graph name
}