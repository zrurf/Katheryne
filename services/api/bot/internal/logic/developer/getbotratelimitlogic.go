package developer

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotRateLimitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotRateLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotRateLimitLogic {
	return &GetBotRateLimitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotRateLimitLogic) GetBotRateLimit(req *types.GetBotRateLimitReq) (resp *types.GetBotRateLimitResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, req.BotID, uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	messagesPerMinute, messagesPerDay, apiCallsPerMinute, err := l.svcCtx.BotDao.GetRateLimit(l.ctx, req.BotID)
	if err != nil {
		return &types.GetBotRateLimitResp{
			MessagesPerMinute: 60,
			MessagesPerDay:    1000,
			ApiCallsPerMinute: 120,
		}, nil
	}

	return &types.GetBotRateLimitResp{
		MessagesPerMinute: messagesPerMinute,
		MessagesPerDay:    messagesPerDay,
		ApiCallsPerMinute: apiCallsPerMinute,
	}, nil
}