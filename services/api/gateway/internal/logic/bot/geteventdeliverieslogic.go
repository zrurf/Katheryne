package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

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
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	result, err := l.svcCtx.BotRpc.GetEventDeliveries(l.ctx, &botclient.GetEventDeliveriesReq{
		BotId:     botId,
		ConvId:    convId,
		EventType: req.EventType,
		Status:    req.Status,
		Page:      int32(req.Page),
		Size:      int32(req.Size),
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.EventDeliveryItem, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, types.EventDeliveryItem{
			EventId:     item.EventId,
			BotId:       req.BotId,
			ConvId:      strconv.FormatInt(item.ConvId, 10),
			EventType:   item.EventType,
			Status:      item.Status,
			RetryCount:  item.RetryCount,
			LastError:   item.LastError,
			DeliveredAt: item.DeliveredAt,
			CreatedAt:   item.CreatedAt,
		})
	}
	return &types.GetEventDeliveriesResp{
		List:  list,
		Total: result.Total,
	}, nil
}
