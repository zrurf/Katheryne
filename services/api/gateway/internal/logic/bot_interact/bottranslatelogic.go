package bot_interact

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotTranslateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotTranslateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotTranslateLogic {
	return &BotTranslateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

const (
	translateCacheTTL = 7 * 24 * time.Hour
)

func translateCacheKey(sourceText, targetLang string) string {
	hash := md5.Sum([]byte(sourceText))
	return fmt.Sprintf("translate:%s:%s", hex.EncodeToString(hash[:]), targetLang)
}

func (l *BotTranslateLogic) BotTranslate(req *types.BotTranslateReq) (resp *types.BotTranslateResp, err error) {
	targetLang := req.TargetLang
	if targetLang == "" {
		targetLang = "zh"
	}

	// 1. Check Redis cache first
	cacheKey := translateCacheKey(req.Text, targetLang)
	if l.svcCtx.Redis != nil {
		cached, cacheErr := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
		if cacheErr == nil && cached != "" {
			// Format: "sourceLang|translatedText"
			parts := strings.SplitN(cached, "|", 2)
			if len(parts) == 2 {
				return &types.BotTranslateResp{
					Text:       parts[1],
					SourceLang: parts[0],
					TargetLang: targetLang,
				}, nil
			}
		}
	}

	// 2. Cache miss — call AI bot
	url := l.svcCtx.Config.AiBotUrl + "/bot/translate"

	body, err := json.Marshal(map[string]interface{}{
		"text":        req.Text,
		"source_lang": req.SourceLang,
		"target_lang": targetLang,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(l.ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		l.Errorf("ai-bot translate request failed: %v", err)
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		l.Errorf("ai-bot translate returned %d: %s", httpResp.StatusCode, string(respBody))
		return nil, fmt.Errorf("ai-bot translate failed: %s", string(respBody))
	}

	// ai-bot wraps response in {"code":0,"msg":"ok","data":{...}}
	var wrapper struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data *types.BotTranslateResp `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.Code != 0 {
		return nil, fmt.Errorf("ai-bot translate error: %s", wrapper.Msg)
	}

	result := wrapper.Data
	if result == nil {
		return nil, fmt.Errorf("ai-bot translate returned empty data")
	}
	result.TargetLang = targetLang

	// 3. Save to Redis cache
	if l.svcCtx.Redis != nil {
		cachedVal := result.SourceLang + "|" + result.Text
		if setErr := l.svcCtx.Redis.Set(l.ctx, cacheKey, cachedVal, translateCacheTTL).Err(); setErr != nil {
			l.Errorf("failed to cache translation: %v", setErr)
		}
	}

	return result, nil
}
