package logic

import (
	"ai-bot/internal/llm"
	"ai-bot/internal/types"
	"strings"

	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type LLMEngine struct {
	provider     llm.Provider
	config       HandlerConfig
	toolExecutor *ToolExecutor
	ragClient    ragclient.Rag
}

func NewLLMEngine(cfg HandlerConfig) *LLMEngine {
	factory := llm.NewFactory()
	provider := factory.Create(cfg.LLMProvider, llm.ProviderConfig{
		APIKey:      cfg.LLMAPIKey,
		BaseURL:     cfg.LLMBaseURL,
		Model:       cfg.LLMModel,
		MaxTokens:   cfg.LLMMaxTokens,
		Temperature: cfg.LLMTemperature,
	})

	return &LLMEngine{
		provider:     provider,
		config:       cfg,
		toolExecutor: NewToolExecutor(cfg.RagClient, cfg.KBIDs),
		ragClient:    cfg.RagClient,
	}
}

func (e *LLMEngine) Chat(messages []types.ChatMessage, systemPrompt string) (string, error) {
	reply, err := e.provider.Chat(messages, systemPrompt, e.config.LLMMaxTokens, e.config.LLMTemperature)
	if err != nil {
		return "", err
	}

	// Execute tool calls if LLM requested any
	if strings.Contains(reply, "<tool_call>") {
		replaced, hasTools := e.toolExecutor.ExecuteToolCalls(nil, reply)
		if hasTools {
			return replaced, nil
		}
	}

	return reply, nil
}

func (e *LLMEngine) SetToolDefinitions(defsJSON string) {
	e.toolExecutor.LoadCustomTools(defsJSON)
}

func (e *LLMEngine) GetToolsDescription() string {
	return e.toolExecutor.GetAvailableToolsDescription()
}

func (e *LLMEngine) GetAvailableTools() []Tool {
	return e.toolExecutor.GetAvailableToolsList()
}

func (e *LLMEngine) Summarize(msgs []types.ChatMessage) (*types.SummarizeResponse, error) {
	convText := buildConversationText(msgs)
	if convText == "" {
		return nil, nil
	}

	systemPrompt := `你是一个专业的聊天记录总结助手。请根据以下对话内容，完成三项任务：
1. 【总结】：用2-3句话概括对话的主要内容
2. 【关键要点】：提取2-3个关键要点
3. 【待办事项】：列出对话中提到的待办事项或行动计划

请以JSON格式返回结果，包含 summary（字符串）、key_points（字符串数组）、action_items（字符串数组）三个字段。只返回JSON，不要有其他内容。`

	userPrompt := "请总结以下对话内容：\n\n" + convText

	userMsgs := []types.ChatMessage{
		{Role: "user", Content: userPrompt},
	}

	result, err := e.provider.Chat(userMsgs, systemPrompt, 2048, 0.3)
	if err != nil {
		return nil, err
	}

	resp, err := parseSummaryJSON(result)
	if err != nil {
		logx.Errorf("parse summary JSON failed: %v, raw: %s", err, result)
		return &types.SummarizeResponse{
			Summary: strings.TrimSpace(result),
		}, nil
	}

	return resp, nil
}

func (e *LLMEngine) SuggestReplies(msgs []types.ChatMessage, count int) (*types.ReplySuggestionResponse, error) {
	if count <= 0 {
		count = 3
	}

	convText := buildConversationText(msgs)
	if convText == "" {
		return nil, nil
	}

	systemPrompt := `你是一个智能聊天助手。根据最近的对话上下文，生成几条自然、得体的回复建议。

要求：
- 每条回复建议应该简短、自然、符合对话上下文
- 回复建议应该是用户可以直接选择使用的完整句子
- 返回3-5条建议

请以JSON格式返回，包含 suggestions 字符串数组。只返回JSON，不要有其他内容。`

	userPrompt := "根据以下对话，生成回复建议：\n\n" + convText

	userMsgs := []types.ChatMessage{
		{Role: "user", Content: userPrompt},
	}

	result, err := e.provider.Chat(userMsgs, systemPrompt, 1024, 0.8)
	if err != nil {
		return nil, err
	}

	parsed, err := parseSuggestionsJSON(result)
	if err != nil {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		var suggestions []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "1. ")
			line = strings.TrimPrefix(line, "2. ")
			line = strings.TrimPrefix(line, "3. ")
			line = strings.TrimPrefix(line, "4. ")
			line = strings.TrimPrefix(line, "5. ")
			suggestions = append(suggestions, line)
		}
		return &types.ReplySuggestionResponse{Suggestions: suggestions}, nil
	}

	return parsed, nil
}

func (e *LLMEngine) Translate(text, sourceLang, targetLang string) (*types.TranslateResponse, error) {
	if targetLang == "" {
		targetLang = "en"
	}

	var prompt string
	if sourceLang != "" {
		prompt = "将以下文本从" + sourceLang + "翻译成" + targetLang + "，只输出翻译结果，不要有任何解释：\n\n" + text
	} else {
		prompt = "将以下文本翻译成" + targetLang + "，自动检测源语言。只输出翻译结果，不要有任何解释：\n\n" + text
	}

	msgs := []types.ChatMessage{
		{Role: "user", Content: prompt},
	}

	result, err := e.provider.Chat(msgs, "", 2048, 0.3)
	if err != nil {
		return nil, err
	}

	return &types.TranslateResponse{
		Text:       strings.TrimSpace(result),
		SourceLang: sourceLang,
		TargetLang: targetLang,
	}, nil
}

func (e *LLMEngine) Moderate(text string) (bool, string, error) {
	systemPrompt := `你是一个内容审核助手。请审核以下消息内容是否包含不当内容（例如色情、暴力、仇恨言论、骚扰、垃圾广告等）。

请以JSON格式返回：{"safe": true/false, "reason": "原因说明"}

只返回JSON，不要有其他内容。`

	msgs := []types.ChatMessage{
		{Role: "user", Content: text},
	}

	result, err := e.provider.Chat(msgs, systemPrompt, 512, 0.0)
	if err != nil {
		return true, "", err
	}

	safe, reason := parseModerationJSON(result)
	return safe, reason, nil
}

func buildConversationText(msgs []types.ChatMessage) string {
	if len(msgs) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString(m.Role)
		sb.WriteString(": ")
		sb.WriteString(m.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}
