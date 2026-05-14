package developer

import (
	"context"

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
	return &types.GetBotRateLimitResp{
		MessagesPerMinute: 60,
		MessagesPerDay:    10000,
		ApiCallsPerMinute: 120,
	}, nil
}