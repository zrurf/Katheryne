package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReadIntervalsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReadIntervalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReadIntervalsLogic {
	return &GetReadIntervalsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReadIntervalsLogic) GetReadIntervals(in *message.GetReadIntervalsReq) (*message.GetReadIntervalsResp, error) {
	intervals, err := l.svcCtx.MessageDao.GetReadIntervals(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	list := make([]*message.ReadIntervalItem, 0, len(intervals))
	for _, iv := range intervals {
		list = append(list, &message.ReadIntervalItem{
			Id:         iv.Id,
			ConvId:     iv.ConvId,
			Uid:        iv.Uid,
			StartMsgId: iv.StartMsgId,
			EndMsgId:   iv.EndMsgId,
			CreatedAt:  iv.CreatedAt.UnixMilli(),
		})
	}

	return &message.GetReadIntervalsResp{List: list}, nil
}