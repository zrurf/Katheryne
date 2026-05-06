// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryEventDeliveryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryEventDeliveryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryEventDeliveryLogic {
	return &RetryEventDeliveryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryEventDeliveryLogic) RetryEventDelivery(req *types.RetryEventDeliveryReq) (resp *types.RetryEventDeliveryResp, err error) {
	// todo: add your logic here and delete this line

	return
}
