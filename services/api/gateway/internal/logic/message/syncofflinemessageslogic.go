// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package message

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncOfflineMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncOfflineMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncOfflineMessagesLogic {
	return &SyncOfflineMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncOfflineMessagesLogic) SyncOfflineMessages(req *types.SyncOfflineMessagesReq) (resp *types.SyncOfflineMessagesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
