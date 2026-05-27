package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	BotAPITokenURL   string
	BotAPIRefreshURL string
	WSGatewayURL     string
	BotClientID      string
	BotClientSecret  string
	LLM              LLMConfig
	RagRpc           zrpc.RpcClientConf
	BotRpc           zrpc.RpcClientConf
	MemRpc           zrpc.RpcClientConf
}

type LLMConfig struct {
	Provider    string  `json:"Provider,default=openai"`
	APIKey      string  `json:"APIKey,env=OPENAI_API_KEY"`
	BaseURL     string  `json:"BaseURL,env=OPENAI_BASE_URL"`
	Model       string  `json:"Model,env=LLM_MODEL"`
	MaxTokens   int     `json:"MaxTokens,default=4096"`
	Temperature float64 `json:"Temperature,default=0.7"`
}
