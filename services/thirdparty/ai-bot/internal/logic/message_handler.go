package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"ai-bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandlerConfig struct {
	LLMProvider    string
	LLMAPIKey      string
	LLMBaseURL     string
	LLMModel       string
	LLMMaxTokens   int
	LLMTemperature float64
}

type BotSender interface {
	SendMessage(data interface{}) error
}

type MessageHandler struct {
	engine *LLMEngine
	sender BotSender
	config HandlerConfig

	convCache   map[string][]types.ChatMessage
	cacheMu     sync.RWMutex
	maxCacheLen int

	// Stats
	totalMessages  int64
	totalTokens    int64
	messagesByConv map[string]int64
	statsMu        sync.RWMutex
	lastActivity   int64
}

func NewMessageHandler(cfg HandlerConfig) *MessageHandler {
	return &MessageHandler{
		engine:         NewLLMEngine(cfg),
		config:         cfg,
		convCache:      make(map[string][]types.ChatMessage),
		maxCacheLen:    50,
		messagesByConv: make(map[string]int64),
	}
}

func (h *MessageHandler) SetSender(sender BotSender) {
	h.sender = sender
}

func (h *MessageHandler) HandleEvent(event *types.EventMessage) {
	switch event.EventType {
	case "message.create":
		h.handleMessageCreate(event)
	case "message.recall":
	case "message.edit":
	case "group.join":
		h.handleGroupJoin(event)
	case "group.leave":
	default:
		logx.Infof("unhandled event type: %s", event.EventType)
	}
}

func (h *MessageHandler) HandleReply(msgID, content string) {
	logx.Infof("received reply: msgID=%s, content=%s", msgID, content)
}

func (h *MessageHandler) handleMessageCreate(event *types.EventMessage) {
	var msgEvent types.MessageCreateEvent
	if err := json.Unmarshal(event.Data, &msgEvent); err != nil {
		logx.Errorf("unmarshal message.create event: %v", err)
		return
	}

	logx.Infof("message.create: convID=%s, sender=%s, content=%s",
		msgEvent.ConvID, msgEvent.SenderName, truncate(msgEvent.Content, 100))

	h.trackMessage(msgEvent.ConvID)

	h.cacheMessage(msgEvent.ConvID, types.ChatMessage{
		Role:    "user",
		Content: msgEvent.SenderName + ": " + msgEvent.Content,
	})

	content := strings.TrimSpace(msgEvent.Content)

	if strings.Contains(content, "@Katheryne") || strings.Contains(content, "@katheryne") {
		h.handleAtBot(&msgEvent, content)
	} else if strings.HasPrefix(content, "/summary") || strings.HasPrefix(content, "/总结") {
		h.handleSummary(msgEvent.ConvID)
	} else if strings.HasPrefix(content, "/translate") || strings.HasPrefix(content, "/翻译") {
		h.handleTranslate(msgEvent.ConvID, content)
	} else if strings.HasPrefix(content, "/moderate") || strings.HasPrefix(content, "/审核") {
		h.handleModerate(msgEvent.ConvID, content)
	} else if strings.HasPrefix(content, "/suggest") || strings.HasPrefix(content, "/建议") {
		h.handleSuggest(msgEvent.ConvID)
	}
}

func (h *MessageHandler) handleAtBot(event *types.MessageCreateEvent, content string) {
	convID := event.ConvID

	prompt := strings.Replace(content, "@Katheryne", "", 1)
	prompt = strings.Replace(prompt, "@katheryne", "", 1)
	prompt = strings.TrimSpace(prompt)

	if prompt == "" {
		prompt = "你好，有什么可以帮助你的？"
	}

	messages := h.getCachedMessages(convID)
	messages = append(messages, types.ChatMessage{
		Role:    "user",
		Content: prompt,
	})

	systemPrompt := `你是 Katheryne，一个友好的智能助手。你是 Katheryne 即时通讯平台的内置 AI 助手。
你可以帮助用户回答问题、总结对话、翻译文本、审核内容等。
你还可以使用以下工具：
- web_search: 搜索互联网获取实时信息（天气、新闻、技术问答等）。当你需要实时信息时，请使用工具调用。

使用工具时，请按以下格式输出:
<tool_call>{"name": "web_search", "args": {"query": "搜索关键词"}}</tool_call>

请用自然、友好的语言回复用户。如果用户使用中文，请用中文回复；如果用户使用其他语言，请用相同语言回复。`

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("panic in handleAtBot: %v", r)
			}
		}()

		reply, err := h.engine.Chat(messages, systemPrompt)
		if err != nil {
			logx.Errorf("AI chat failed: %v", err)
			h.sendTextMessage(convID, "抱歉，我暂时无法回答你的问题，请稍后再试。")
			return
		}

		h.trackTokens(int64(len(reply)))

		h.sendTextMessage(convID, reply)

		h.cacheMessage(convID, types.ChatMessage{
			Role:    "assistant",
			Content: reply,
		})
	}()
}

func (h *MessageHandler) handleSummary(convID string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("panic in handleSummary: %v", r)
			}
		}()

		messages := h.getCachedMessages(convID)
		if len(messages) == 0 {
			h.sendTextMessage(convID, "暂无对话内容可以总结。")
			return
		}

		h.sendTextMessage(convID, "正在生成对话总结...")

		result, err := h.engine.Summarize(messages)
		if err != nil {
			logx.Errorf("summary failed: %v", err)
			h.sendTextMessage(convID, "生成总结时出错，请稍后再试。")
			return
		}

		var sb strings.Builder
		sb.WriteString("**对话总结**\n\n")

		if result.Summary != "" {
			sb.WriteString("📝 **总结**\n")
			sb.WriteString(result.Summary)
			sb.WriteString("\n\n")
		}

		if len(result.KeyPoints) > 0 {
			sb.WriteString("🔑 **关键要点**\n")
			for _, kp := range result.KeyPoints {
				sb.WriteString(fmt.Sprintf("- %s\n", kp))
			}
			sb.WriteString("\n")
		}

		if len(result.ActionItems) > 0 {
			sb.WriteString("✅ **待办事项**\n")
			for _, ai := range result.ActionItems {
				sb.WriteString(fmt.Sprintf("- [ ] %s\n", ai))
			}
		}

		h.sendTextMessage(convID, sb.String())
	}()
}

func (h *MessageHandler) handleTranslate(convID, content string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("panic in handleTranslate: %v", r)
			}
		}()

		text := strings.TrimPrefix(content, "/translate")
		text = strings.TrimPrefix(text, "/翻译")
		text = strings.TrimSpace(text)

		targetLang := "en"
		parts := strings.SplitN(text, " ", 2)
		if len(parts) == 2 && len(parts[0]) <= 5 {
			targetLang = parts[0]
			text = parts[1]
		}

		if text == "" {
			h.sendTextMessage(convID, "用法：/translate [目标语言] <文本>")
			return
		}

		result, err := h.engine.Translate(text, "", targetLang)
		if err != nil {
			logx.Errorf("translate failed: %v", err)
			h.sendTextMessage(convID, "翻译失败，请稍后再试。")
			return
		}

		h.sendTextMessage(convID, fmt.Sprintf("🌐 **翻译结果** (%s → %s)\n\n%s", result.SourceLang, result.TargetLang, result.Text))
	}()
}

func (h *MessageHandler) handleModerate(convID, content string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("panic in handleModerate: %v", r)
			}
		}()

		text := strings.TrimPrefix(content, "/moderate")
		text = strings.TrimPrefix(text, "/审核")
		text = strings.TrimSpace(text)

		if text == "" {
			h.sendTextMessage(convID, "用法：/moderate <文本>")
			return
		}

		safe, reason, err := h.engine.Moderate(text)
		if err != nil {
			logx.Errorf("moderate failed: %v", err)
			h.sendTextMessage(convID, "审核失败，请稍后再试。")
			return
		}

		if safe {
			h.sendTextMessage(convID, "✅ 内容审核通过，未检测到不当内容。")
		} else {
			h.sendTextMessage(convID, fmt.Sprintf("⚠️ 内容审核不通过\n\n原因：%s", reason))
		}
	}()
}

func (h *MessageHandler) handleSuggest(convID string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("panic in handleSuggest: %v", r)
			}
		}()

		messages := h.getCachedMessages(convID)
		result, err := h.engine.SuggestReplies(messages, 3)
		if err != nil {
			logx.Errorf("suggest replies failed: %v", err)
			h.sendTextMessage(convID, "生成回复建议失败，请稍后再试。")
			return
		}

		if result == nil || len(result.Suggestions) == 0 {
			h.sendTextMessage(convID, "暂无回复建议。")
			return
		}

		var sb strings.Builder
		sb.WriteString("💡 **回复建议**\n\n")
		for i, s := range result.Suggestions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, s))
		}
		h.sendTextMessage(convID, sb.String())
	}()
}

func (h *MessageHandler) handleGroupJoin(event *types.EventMessage) {
	var data struct {
		ConvID     string `json:"conv_id"`
		GroupName  string `json:"group_name"`
		MemberName string `json:"member_name"`
	}
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	h.sendTextMessage(data.ConvID, fmt.Sprintf("👋 欢迎 @%s 加入群聊「%s」！我是 Katheryne AI，有什么问题可以 @我。",
		data.MemberName, data.GroupName))
}

func (h *MessageHandler) sendTextMessage(convID, content string) {
	if h.sender == nil {
		logx.Errorf("sender not set, cannot send message")
		return
	}

	msg := map[string]interface{}{
		"type": "send_message",
		"data": map[string]interface{}{
			"conv_id":      convID,
			"msg_type":     "TEXT",
			"content":      content,
			"content_type": "MARKDOWN",
		},
	}

	if err := h.sender.SendMessage(msg); err != nil {
		logx.Errorf("send message failed: %v", err)
	}
}

func (h *MessageHandler) cacheMessage(convID string, msg types.ChatMessage) {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	cache, ok := h.convCache[convID]
	if !ok {
		cache = make([]types.ChatMessage, 0, h.maxCacheLen)
	}

	cache = append(cache, msg)

	if len(cache) > h.maxCacheLen {
		cache = cache[len(cache)-h.maxCacheLen:]
	}

	h.convCache[convID] = cache
}

func (h *MessageHandler) getCachedMessages(convID string) []types.ChatMessage {
	h.cacheMu.RLock()
	defer h.cacheMu.RUnlock()

	cache, ok := h.convCache[convID]
	if !ok {
		return nil
	}

	result := make([]types.ChatMessage, len(cache))
	copy(result, cache)
	return result
}

func (h *MessageHandler) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.cleanExpiredCache()
			}
		}
	}()
}

func (h *MessageHandler) cleanExpiredCache() {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	for k := range h.convCache {
		if len(h.convCache[k]) == 0 {
			delete(h.convCache, k)
		}
	}
}

// GetStatus returns the current connection status of the bot
func (h *MessageHandler) GetStatus() string {
	return "connected"
}

// ReloadConfig reloads the LLM engine with new configuration
func (h *MessageHandler) ReloadConfig(cfg HandlerConfig) {
	h.config.LLMProvider = cfg.LLMProvider
	h.config.LLMAPIKey = cfg.LLMAPIKey
	h.config.LLMBaseURL = cfg.LLMBaseURL
	h.config.LLMModel = cfg.LLMModel
	h.config.LLMMaxTokens = cfg.LLMMaxTokens
	h.config.LLMTemperature = cfg.LLMTemperature
	h.engine = NewLLMEngine(h.config)
	logx.Infof("Bot LLM config reloaded: provider=%s, model=%s", cfg.LLMProvider, cfg.LLMModel)
}

// GetStats returns bot usage statistics
func (h *MessageHandler) GetStats() map[string]interface{} {
	h.statsMu.RLock()
	defer h.statsMu.RUnlock()

	h.cacheMu.RLock()
	activeConvs := len(h.convCache)
	h.cacheMu.RUnlock()

	return map[string]interface{}{
		"total_messages": h.totalMessages,
		"total_tokens":   h.totalTokens,
		"active_convs":   activeConvs,
		"last_activity":  h.lastActivity,
	}
}

// GetMemory returns conversation memory for a specific conversation or all
func (h *MessageHandler) GetMemory(convID string) map[string][]types.ChatMessage {
	h.cacheMu.RLock()
	defer h.cacheMu.RUnlock()

	if convID != "" {
		messages, ok := h.convCache[convID]
		if !ok {
			return nil
		}
		result := make(map[string][]types.ChatMessage)
		result[convID] = messages
		return result
	}

	result := make(map[string][]types.ChatMessage)
	for k, v := range h.convCache {
		result[k] = v
	}
	return result
}

// ClearMemory clears conversation memory
func (h *MessageHandler) ClearMemory(convID string) {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	if convID != "" {
		delete(h.convCache, convID)
	} else {
		h.convCache = make(map[string][]types.ChatMessage)
	}
}

// trackMessage increments message stats
func (h *MessageHandler) trackMessage(convID string) {
	h.statsMu.Lock()
	defer h.statsMu.Unlock()
	h.totalMessages++
	h.messagesByConv[convID]++
	h.lastActivity = time.Now().Unix()
}

func (h *MessageHandler) trackTokens(count int64) {
	h.statsMu.Lock()
	defer h.statsMu.Unlock()
	h.totalTokens += count
}

// Public Bot API methods for REST endpoints

func (h *MessageHandler) SummarizeContext(messages []types.ChatMessage) (*types.SummarizeResponse, error) {
	return h.engine.Summarize(messages)
}

func (h *MessageHandler) TranslateText(text, sourceLang, targetLang string) (*types.TranslateResponse, error) {
	return h.engine.Translate(text, sourceLang, targetLang)
}

func (h *MessageHandler) SuggestRepliesContext(messages []types.ChatMessage) (*types.ReplySuggestionResponse, error) {
	return h.engine.SuggestReplies(messages, 3)
}

func (h *MessageHandler) ModerateText(text string) (*types.ModerateResponse, error) {
	safe, reason, err := h.engine.Moderate(text)
	if err != nil {
		return nil, err
	}
	return &types.ModerateResponse{Safe: safe, Reason: reason}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
