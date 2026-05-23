package llm

import (
	"ai-bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type Provider interface {
	Chat(messages []types.ChatMessage, systemPrompt string, maxTokens int, temperature float64) (string, error)
	ChatStream(messages []types.ChatMessage, systemPrompt string, maxTokens int, temperature float64, callback func(chunk string)) error
}

type Factory struct {
	providers map[string]func(config ProviderConfig) Provider
}

type ProviderConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
}

func NewFactory() *Factory {
	f := &Factory{
		providers: make(map[string]func(config ProviderConfig) Provider),
	}
	f.Register("openai", func(cfg ProviderConfig) Provider {
		return NewOpenAIProvider(cfg)
	})
	f.Register("anthropic", func(cfg ProviderConfig) Provider {
		return NewAnthropicProvider(cfg)
	})
	f.Register("claude", func(cfg ProviderConfig) Provider {
		return NewAnthropicProvider(cfg)
	})
	return f
}

func (f *Factory) Register(name string, factory func(config ProviderConfig) Provider) {
	f.providers[name] = factory
}

func (f *Factory) Create(name string, config ProviderConfig) Provider {
	factory, ok := f.providers[name]
	if !ok {
		logx.Errorf("unknown LLM provider: %s, falling back to openai", name)
		factory = f.providers["openai"]
	}
	return factory(config)
}