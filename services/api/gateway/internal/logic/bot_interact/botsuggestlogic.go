package bot_interact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotSuggestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotSuggestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotSuggestLogic {
	return &BotSuggestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotSuggestLogic) BotSuggest(req *types.BotSuggestReq) (resp *types.BotSuggestResp, err error) {
	convID, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		l.Errorf("invalid conv_id: %s", req.ConvID)
		return nil, err
	}

	// Fetch recent messages for context
	msgs, err := l.svcCtx.MessageRpc.GetMessages(l.ctx, &messageclient.GetMessagesReq{
		ConvId: convID,
		Limit:  20,
	})
	if err != nil {
		l.Errorf("GetMessages RPC failed: %v", err)
		return nil, err
	}

	// Build chat messages for ai-bot
	chatMsgs := make([]map[string]string, 0, len(msgs.List))
	for _, m := range msgs.List {
		role := "user"
		chatMsgs = append(chatMsgs, map[string]string{
			"role":    role,
			"content": m.Content,
		})
	}

	url := l.svcCtx.Config.AiBotUrl + "/bot/suggest"

	body, err := json.Marshal(map[string]interface{}{
		"messages": chatMsgs,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(l.ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		l.Errorf("ai-bot suggest request failed: %v", err)
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		l.Errorf("ai-bot suggest returned %d: %s", httpResp.StatusCode, string(respBody))
		return nil, fmt.Errorf("ai-bot suggest failed: %s", string(respBody))
	}

	var wrapper struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data *types.BotSuggestResp `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.Code != 0 {
		return nil, fmt.Errorf("ai-bot suggest error: %s", wrapper.Msg)
	}
	if wrapper.Data == nil {
		return &types.BotSuggestResp{}, nil
	}

	return wrapper.Data, nil
}
