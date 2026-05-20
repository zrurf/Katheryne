package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEventDeliveriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEventDeliveriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventDeliveriesLogic {
	return &GetEventDeliveriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetEventDeliveriesLogic) GetEventDeliveries(in *bot.GetEventDeliveriesReq) (*bot.GetEventDeliveriesResp, error) {
	// uid is not directly in request for this endpoint; ownership is checked via bot_id
	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, in.BotId, 0); err != nil {
		// The request doesn't have uid, so we'll do ownership check inside the DAO
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	list, total, err := l.svcCtx.EventDao.ListEventDeliveries(l.ctx, in.BotId, in.ConvId, in.EventType, in.Status, in.Page, in.Size)
	if err != nil {
		return nil, err
	}

	return &bot.GetEventDeliveriesResp{
		List:  list,
		Total: total,
	}, nil
}