package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	BotAPITokenURL   string
	BotAPIRefreshURL string
	WSGatewayURL     string
	BotClientID      string
	BotClientSecret  string
	LLM              LLMConfig
}

type LLMConfig struct {
	Provider    string  `json:"Provider,default=openai"`
	APIKey      string  `json:"APIKey"`
	BaseURL     string  `json:"BaseURL,default=https://api.openai.com/v1"`
	Model       string  `json:"Model,default=gpt-4o"`
	MaxTokens   int     `json:"MaxTokens,default=4096"`
	Temperature float64 `json:"Temperature,default=0.7"`
}