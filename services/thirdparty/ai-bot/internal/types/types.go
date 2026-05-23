package types

import (
	"encoding/json"
	"time"
)

type BotCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	BotID        int64  `json:"bot_id"`
}

type TokenCache struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type WSMessage struct {
	Op   string          `json:"op"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
	Seq  int64           `json:"seq,omitempty"`
}

type EventMessage struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	ConvID    string          `json:"conv_id"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type MessageCreateEvent struct {
	MsgID       string `json:"msg_id"`
	ConvID      string `json:"conv_id"`
	SenderUID   string `json:"sender_uid"`
	SenderName  string `json:"sender_name"`
	MsgType     string `json:"msg_type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	QuoteMsgID  string `json:"quote_msg_id,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

type BotSendMessageData struct {
	ConvID      string `json:"conv_id"`
	MsgType     string `json:"msg_type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
	QuoteMsgID  string `json:"quote_msg_id,omitempty"`
}

type SummarizeRequest struct {
	ConvID    string        `json:"conv_id"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type SummarizeResponse struct {
	Summary     string   `json:"summary"`
	KeyPoints   []string `json:"key_points"`
	ActionItems []string `json:"action_items"`
}

type ReplySuggestionRequest struct {
	ConvID   string        `json:"conv_id"`
	Messages []ChatMessage `json:"messages"`
	Count    int           `json:"count,omitempty"`
}

type ReplySuggestionResponse struct {
	Suggestions []string `json:"suggestions"`
}

type TranslateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang,omitempty"`
	TargetLang string `json:"target_lang"`
}

type TranslateResponse struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMConfig struct {
	Provider    string  `json:"provider"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type ModerateResponse struct {
	Safe   bool   `json:"safe"`
	Reason string `json:"reason"`
}
