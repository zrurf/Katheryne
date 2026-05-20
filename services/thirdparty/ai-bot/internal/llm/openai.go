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

type OpenAIProvider struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
}

func NewOpenAIProvider(cfg ProviderConfig) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:      cfg.APIKey,
		baseURL:     cfg.BaseURL,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type chatCompletionReq struct {
	Model       string                      `json:"model"`
	Messages    []chatMessage               `json:"messages"`
	MaxTokens   int                         `json:"max_tokens,omitempty"`
	Temperature float64                     `json:"temperature,omitempty"`
	Stream      bool                        `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) Chat(messages []types.ChatMessage, systemPrompt string, maxTokens int, temperature float64) (string, error) {
	if temperature == 0 {
		temperature = p.temperature
	}
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}

	msgs := make([]chatMessage, 0, len(messages)+1)
	if systemPrompt != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: systemPrompt})
	}
	for _, m := range messages {
		msgs = append(msgs, chatMessage{Role: m.Role, Content: m.Content})
	}

	reqBody := chatCompletionReq{
		Model:       p.model,
		Messages:    msgs,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var chatResp chatCompletionResp
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) ChatStream(messages []types.ChatMessage, systemPrompt string, maxTokens int, temperature float64, callback func(chunk string)) error {
	if temperature == 0 {
		temperature = p.temperature
	}
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}

	msgs := make([]chatMessage, 0, len(messages)+1)
	if systemPrompt != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: systemPrompt})
	}
	for _, m := range messages {
		msgs = append(msgs, chatMessage{Role: m.Role, Content: m.Content})
	}

	reqBody := chatCompletionReq{
		Model:       p.model,
		Messages:    msgs,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      true,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			logx.Errorf("stream decode error: %v", err)
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			callback(chunk.Choices[0].Delta.Content)
		}
	}

	return nil
}