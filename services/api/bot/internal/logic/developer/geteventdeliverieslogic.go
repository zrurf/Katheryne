package developer

import (
	"context"

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
	return &types.GetEventDeliveriesResp{
		List:  []types.EventDeliveryItem{},
		Total: 0,
	}, nil
}