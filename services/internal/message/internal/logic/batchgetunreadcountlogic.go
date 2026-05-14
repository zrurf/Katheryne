package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetUnreadCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUnreadCountLogic {
	return &BatchGetUnreadCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchGetUnreadCountLogic) BatchGetUnreadCount(in *message.BatchGetUnreadCountReq) (*message.BatchGetUnreadCountResp, error) {
	result, err := l.svcCtx.MessageDao.BatchGetUnreadCount(l.ctx, in.Uid, in.ConvIds)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	list := make([]*message.ConvUnread, 0, len(result))
	for convId, count := range result {
		list = append(list, &message.ConvUnread{
			ConvId: convId,
			Count:  count,
		})
	}

	return &message.BatchGetUnreadCountResp{List: list}, nil
}
