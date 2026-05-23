package bot_interact

import (
	"context"
	"encoding/json"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotSummarizeResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotSummarizeResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotSummarizeResultLogic {
	return &BotSummarizeResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotSummarizeResultLogic) BotSummarizeResult(req *types.SummarizeResultReq) (resp *types.SummarizeResultResp, err error) {
	if req.Ticket == "" {
		return &types.SummarizeResultResp{Status: "error", Error: "missing ticket"}, nil
	}

	val, err := l.svcCtx.Redis.Get(l.ctx, "summarize:"+req.Ticket).Result()
	if err != nil {
		// Key expired or never existed
		return &types.SummarizeResultResp{Status: "error", Error: "ticket not found or expired"}, nil
	}

	var result struct {
		Status      string   `json:"status"`
		Summary     string   `json:"summary"`
		KeyPoints   []string `json:"key_points"`
		ActionItems []string `json:"action_items"`
		Error       string   `json:"error"`
	}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return &types.SummarizeResultResp{Status: "error", Error: "invalid result format"}, nil
	}

	if result.Status == "completed" {
		return &types.SummarizeResultResp{
			Status: "completed",
			Result: &types.BotSummarizeResp{
				Summary:     result.Summary,
				KeyPoints:   result.KeyPoints,
				ActionItems: result.ActionItems,
			},
		}, nil
	}

	if result.Status == "error" {
		return &types.SummarizeResultResp{Status: "error", Error: result.Error}, nil
	}

	return &types.SummarizeResultResp{Status: "processing"}, nil
}
