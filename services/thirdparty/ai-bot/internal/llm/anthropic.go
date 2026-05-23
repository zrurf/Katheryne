package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai-bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnthropicProvider struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
}

func NewAnthropicProvider(cfg ProviderConfig) *AnthropicProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &AnthropicProvider{
		apiKey:      cfg.APIKey,
		baseURL:     baseURL,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type anthropicMessageReq struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Stream      bool               `json:"stream"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessageResp struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *AnthropicProvider) Chat(messages []types.ChatMessage, systemPrompt string, maxTokens int, temperature float64) (string, error) {
	if temperature == 0 {
		temperature = p.temperature
	}
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}

	msgs := make([]anthropicMessage, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	reqBody := anthropicMessageReq{
		Model:       p.model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		System:      systemPrompt,
		Messages:    msgs,
		Stream:      false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var result anthropicMessageResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return result.Content[0].Text, nil
}

func (p *AnthropicProvider) ChatStream(messages []types.ChatMessage, systemPrompt string, maxTokens int, temperature float64, callback func(chunk string)) error {
	logx.Info("ChatStream not implemented for Anthropic, falling back to Chat")
	reply, err := p.Chat(messages, systemPrompt, maxTokens, temperature)
	if err != nil {
		return err
	}
	callback(reply)
	return nil
}