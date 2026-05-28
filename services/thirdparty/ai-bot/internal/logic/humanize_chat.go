package logic

import (
	"math/rand/v2"
	"strings"
)

// HumanizeChat transforms a bot response to feel more human-like.
// Key characteristics:
//   - Removes trailing periods (真人很少打句号)
//   - Splits long messages into multiple segments
//   - Adds occasional emojis and standalone expressions
//   - Varies message timing for natural feel
func HumanizeChat(content string) []HumanizedSegment {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	// Split into natural message segments
	texts := splitIntoMessages(content)
	if len(texts) == 0 {
		return []HumanizedSegment{{Text: content, DelayMs: 0}}
	}

	var segments []HumanizedSegment
	for i, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		// Remove trailing period from each segment (真人风格)
		text = strings.TrimRight(text, "。.；;，,！!？?")

		// Add occasional emoji or expression
		text = addHumanFlavor(text, i == 0)

		delay := calculateDelay(text)
		segments = append(segments, HumanizedSegment{
			Text:    text,
			DelayMs: delay,
		})
	}

	return segments
}

// HumanizedSegment represents one message segment with a send delay.
type HumanizedSegment struct {
	Text    string
	DelayMs int64
}

// splitIntoMessages splits LLM output into natural message segments.
// Uses paragraph breaks (double newline) as primary split points,
// with fallback to sentence boundaries for very long paragraphs.
func splitIntoMessages(text string) []string {
	// First try splitting by paragraph
	paragraphs := strings.Split(text, "\n\n")
	if len(paragraphs) <= 1 {
		// Try single newlines
		paragraphs = strings.Split(text, "\n")
	}

	if len(paragraphs) <= 1 {
		// Single paragraph — split into sentence-like chunks
		return splitSentenceChunks(text, 80, 300)
	}

	var result []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len([]rune(p)) > 300 {
			// Too long, further split
			result = append(result, splitSentenceChunks(p, 60, 250)...)
		} else {
			result = append(result, p)
		}
	}

	return result
}

// splitSentenceChunks splits text at sentence boundaries into readable chunks.
func splitSentenceChunks(text string, minLen, maxLen int) []string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return []string{text}
	}

	// Find sentence boundaries: 。！？. ! ? followed by space or end
	var chunks []string
	start := 0
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		isBreak := ch == '。' || ch == '！' || ch == '？' || ch == '.' || ch == '!' || ch == '?' || ch == '\n'
		isNewline := ch == '\n'

		// At a break point, check if current chunk is long enough
		if (isBreak || isNewline) && i-start >= minLen {
			// Include the punctuation
			end := i + 1
			if end > len(runes) {
				end = len(runes)
			}
			chunks = append(chunks, string(runes[start:end]))
			start = end
		} else if isNewline {
			// Force split at newline if past minLen
			chunks = append(chunks, string(runes[start:i]))
			start = i + 1
		}
	}

	// Remaining text
	if start < len(runes) {
		remaining := string(runes[start:])
		if len([]rune(remaining)) > maxLen {
			// Hard split by comma boundaries
			chunks = append(chunks, splitByComma(remaining, maxLen)...)
		} else {
			chunks = append(chunks, remaining)
		}
	}

	return chunks
}

// splitByComma splits at comma/顿号 boundaries.
func splitByComma(text string, maxLen int) []string {
	runes := []rune(text)
	var chunks []string
	start := 0
	for i := 0; i < len(runes); i++ {
		if (runes[i] == '，' || runes[i] == '、') && i-start >= maxLen/2 {
			chunks = append(chunks, string(runes[start:i+1]))
			start = i + 1
		}
		if i-start >= maxLen {
			chunks = append(chunks, string(runes[start:i]))
			start = i
		}
	}
	if start < len(runes) {
		chunks = append(chunks, string(runes[start:]))
	}
	return chunks
}

// addHumanFlavor adds random human-like touches to text.
func addHumanFlavor(text string, isFirst bool) string {
	r := rand.Float64()

	// 15% chance to add emoji at the end
	if r < 0.15 {
		emoji := randomEmoji(text)
		if emoji != "" {
			text = text + " " + emoji
		}
	}

	// 8% chance to add a casual prefix
	if isFirst && r > 0.85 && r < 0.93 {
		prefixes := []string{"hmm ", "emmm ", "嗯… ", "啊 ", "哦 "}
		text = prefixes[rand.IntN(len(prefixes))] + text
	}

	return text
}

// randomEmoji returns a context-appropriate emoji.
func randomEmoji(text string) string {
	// Positive/question emojis
	positive := []string{"😊", "🤔", "😄", "😅", "👍", "✨", "💡", "😌", "😂", "🤣"}
	// Simple ones
	simple := []string{"w", "haha", "hmm", "诶", "啧"}

	all := append(positive, simple...)
	return all[rand.IntN(len(all))]
}

// randomStandaloneExpression returns a standalone chat expression.
func randomStandaloneExpression() string {
	expressions := []string{
		"😊", "😄", "🤔", "嗯", "哈哈", "懂了",
		"好嘞", "没问题", "对的", "确实",
		"😅", "😂", "👍", "👌", "ok",
		"了解了", "明白", "收到", "有意思",
	}
	return expressions[rand.IntN(len(expressions))]
}

// calculateDelay computes a natural-seeming delay based on message length.
func calculateDelay(text string) int64 {
	baseDelay := int64(800)
	// ~50ms per character of reading time
	charDelay := int64(len([]rune(text))) * 50
	// Add some randomness
	jitter := randRange(-200, 200)

	total := baseDelay + charDelay + jitter
	if total < 400 {
		total = 400
	}
	if total > 5000 {
		total = 5000
	}
	return total
}

func randRange(min, max int64) int64 {
	return min + rand.Int64N(max-min+1)
}
