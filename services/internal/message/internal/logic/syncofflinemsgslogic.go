package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncOfflineMsgsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncOfflineMsgsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncOfflineMsgsLogic {
	return &SyncOfflineMsgsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SyncOfflineMsgsLogic) SyncOfflineMsgs(in *message.SyncOfflineMsgsReq) (*message.SyncOfflineMsgsResp, error) {
	// todo: add your logic here and delete this line

	return &message.SyncOfflineMsgsResp{}, nil
}
