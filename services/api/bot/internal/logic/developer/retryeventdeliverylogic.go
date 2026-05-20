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
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.EventDao.RetryEvent(l.ctx, req.EventID, uid); err != nil {
		return nil, err
	}

	return &types.RetryEventDeliveryResp{}, nil
}
