package bot_interact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotModerateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotModerateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotModerateLogic {
	return &BotModerateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotModerateLogic) BotModerate(req *types.BotModerateReq) (resp *types.BotModerateResp, err error) {
	url := l.svcCtx.Config.AiBotUrl + "/bot/moderate"

	body, err := json.Marshal(map[string]interface{}{
		"text": req.Text,
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
		l.Errorf("ai-bot moderate request failed: %v", err)
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		l.Errorf("ai-bot moderate returned %d: %s", httpResp.StatusCode, string(respBody))
		return nil, fmt.Errorf("ai-bot moderate failed: %s", string(respBody))
	}

	var wrapper struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data *types.BotModerateResp `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.Code != 0 {
		return nil, fmt.Errorf("ai-bot moderate error: %s", wrapper.Msg)
	}

	return wrapper.Data, nil
}
