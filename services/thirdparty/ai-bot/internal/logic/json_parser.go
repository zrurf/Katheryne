package logic

import (
	"encoding/json"
	"strings"

	"ai-bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

func parseSummaryJSON(raw string) (*types.SummarizeResponse, error) {
	raw = extractJSON(raw)
	var resp types.SummarizeResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func parseSuggestionsJSON(raw string) (*types.ReplySuggestionResponse, error) {
	raw = extractJSON(raw)
	var resp types.ReplySuggestionResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func parseModerationJSON(raw string) (bool, string) {
	raw = extractJSON(raw)
	var resp struct {
		Safe   bool   `json:"safe"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		logx.Errorf("parse moderation JSON failed: %v", err)
		return true, ""
	}
	return resp.Safe, resp.Reason
}

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}

	start = strings.Index(raw, "[")
	end = strings.LastIndex(raw, "]")
	if start >= 0 && end > start {
		return `{"suggestions": ` + raw[start:end+1] + `}`
	}

	return raw
}