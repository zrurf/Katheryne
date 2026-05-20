package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotRateLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBotRateLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotRateLimitLogic {
	return &UpdateBotRateLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateBotRateLimitLogic) UpdateBotRateLimit(in *bot.UpdateBotRateLimitReq) (*bot.UpdateBotRateLimitResp, error) {
	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, in.BotId, in.Uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	updates := make(map[string]int)
	if in.MessagesPerMinute > 0 {
		updates["messages_per_minute"] = int(in.MessagesPerMinute)
	}
	if in.MessagesPerDay > 0 {
		updates["messages_per_day"] = int(in.MessagesPerDay)
	}
	if in.ApiCallsPerMinute > 0 {
		updates["api_calls_per_minute"] = int(in.ApiCallsPerMinute)
	}

	if len(updates) == 0 {
		return &bot.UpdateBotRateLimitResp{}, nil
	}

	if err := l.svcCtx.BotDao.UpdateRateLimit(l.ctx, in.BotId, updates); err != nil {
		return nil, err
	}

	return &bot.UpdateBotRateLimitResp{}, nil
}