package logic

import (
	"fmt"
	"strings"

	"ai-bot/internal/types"
)

const (
	maxContextMessages   = 40
	compressRetainRecent = 20
	maxMsgContentLen     = 2000
)

// compressMessages reduces context size by summarizing older messages.
// Keeps the most recent messages intact, compresses the rest into a summary.
func compressMessages(messages []types.ChatMessage) []types.ChatMessage {
	if len(messages) <= maxContextMessages {
		return truncateMessages(messages)
	}

	splitPoint := len(messages) - compressRetainRecent

	var summary strings.Builder
	summary.WriteString("[对话历史摘要]\n")
	for i := 0; i < splitPoint; i++ {
		msg := messages[i]
		role := "用户"
		if msg.Role == "assistant" {
			role = "AI"
		}
		content := msg.Content
		if len([]rune(content)) > 150 {
			content = string([]rune(content)[:150]) + "..."
		}
		summary.WriteString(fmt.Sprintf("%s: %s\n", role, content))
	}

	compressed := make([]types.ChatMessage, 1+len(messages)-splitPoint)
	compressed[0] = types.ChatMessage{
		Role:    "system",
		Content: summary.String(),
	}
	for i := splitPoint; i < len(messages); i++ {
		compressed[1+i-splitPoint] = types.ChatMessage{
			Role:    messages[i].Role,
			Content: truncateMsgContent(messages[i].Content),
		}
	}

	return compressed
}

func truncateMessages(messages []types.ChatMessage) []types.ChatMessage {
	result := make([]types.ChatMessage, len(messages))
	for i, m := range messages {
		result[i] = types.ChatMessage{
			Role:    m.Role,
			Content: truncateMsgContent(m.Content),
		}
	}
	return result
}

func truncateMsgContent(content string) string {
	runes := []rune(content)
	if len(runes) <= maxMsgContentLen {
		return content
	}
	return string(runes[:maxMsgContentLen]) + "..."
}

// estimatedTokens roughly estimates token count from rune count.
// Approximation: 1 token ≈ 0.75 Chinese characters or 3 English chars (aligned to ~4 chars)
func estimatedTokens(text string) int {
	runes := []rune(text)
	total := 0
	for _, r := range runes {
		if r < 128 {
			total += 1
		} else {
			total += 2
		}
	}
	// Roughly 3-4 input bytes per token for mixed text
	return total / 3
}

func compressMemories(memCtx string) string {
	if memCtx == "" {
		return ""
	}
	lines := strings.Split(memCtx, "\n")
	var compact strings.Builder
	compact.WriteString("[历史记忆]\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "##") {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		compact.WriteString(line)
		compact.WriteString("\n")
	}
	return compact.String()
}