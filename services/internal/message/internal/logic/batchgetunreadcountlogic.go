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
	// todo: add your logic here and delete this line

	return &message.BatchGetUnreadCountResp{}, nil
}
