package developer

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotRateLimitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBotRateLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotRateLimitLogic {
	return &UpdateBotRateLimitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBotRateLimitLogic) UpdateBotRateLimit(req *types.UpdateBotRateLimitReq) (resp *types.UpdateBotRateLimitResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, req.BotID, uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	updates := make(map[string]int)
	if req.MessagesPerMinute > 0 {
		updates["messages_per_minute"] = req.MessagesPerMinute
	}
	if req.MessagesPerDay > 0 {
		updates["messages_per_day"] = req.MessagesPerDay
	}
	if req.ApiCallsPerMinute > 0 {
		updates["api_calls_per_minute"] = req.ApiCallsPerMinute
	}

	if len(updates) == 0 {
		return &types.UpdateBotRateLimitResp{}, nil
	}

	if err := l.svcCtx.BotDao.UpdateRateLimit(l.ctx, req.BotID, updates); err != nil {
		return nil, err
	}

	return &types.UpdateBotRateLimitResp{}, nil
}