package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotRateLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBotRateLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotRateLimitLogic {
	return &GetBotRateLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBotRateLimitLogic) GetBotRateLimit(in *bot.GetBotRateLimitReq) (*bot.GetBotRateLimitResp, error) {
	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, in.BotId, in.Uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	messagesPerMinute, messagesPerDay, apiCallsPerMinute, err := l.svcCtx.BotDao.GetRateLimit(l.ctx, in.BotId)
	if err != nil {
		return &bot.GetBotRateLimitResp{
			MessagesPerMinute: 60,
			MessagesPerDay:    1000,
			ApiCallsPerMinute: 120,
		}, nil
	}

	return &bot.GetBotRateLimitResp{
		MessagesPerMinute: messagesPerMinute,
		MessagesPerDay:    messagesPerDay,
		ApiCallsPerMinute: apiCallsPerMinute,
	}, nil
}