// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

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
	// todo: add your logic here and delete this line

	return
}
