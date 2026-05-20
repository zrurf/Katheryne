package developer

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEventDeliveriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetEventDeliveriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventDeliveriesLogic {
	return &GetEventDeliveriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEventDeliveriesLogic) GetEventDeliveries(req *types.GetEventDeliveriesReq) (resp *types.GetEventDeliveriesResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, req.BotID, uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 || size > 100 {
		size = 20
	}

	list, total, err := l.svcCtx.EventDao.ListEventDeliveries(l.ctx, req.BotID, req.ConvID, req.EventType, req.Status, page, size)
	if err != nil {
		return nil, err
	}

	return &types.GetEventDeliveriesResp{
		List:  list,
		Total: total,
	}, nil
}