// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
