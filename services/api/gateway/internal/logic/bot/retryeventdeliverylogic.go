package bot

import (
	"context"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

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
	_, err = l.svcCtx.BotRpc.RetryEventDelivery(l.ctx, &botclient.RetryEventDeliveryReq{
		EventId: req.EventId,
	})
	if err != nil {
		return nil, err
	}
	return &types.RetryEventDeliveryResp{}, nil
}
