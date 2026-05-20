package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryEventDeliveryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryEventDeliveryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryEventDeliveryLogic {
	return &RetryEventDeliveryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RetryEventDeliveryLogic) RetryEventDelivery(in *bot.RetryEventDeliveryReq) (*bot.RetryEventDeliveryResp, error) {
	if err := l.svcCtx.EventDao.RetryEvent(l.ctx, in.EventId, 0); err != nil {
		return nil, err
	}

	return &bot.RetryEventDeliveryResp{}, nil
}