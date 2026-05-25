// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerExternalSyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTriggerExternalSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerExternalSyncLogic {
	return &TriggerExternalSyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TriggerExternalSyncLogic) TriggerExternalSync(req *types.TriggerExternalSyncRequest) (resp *types.TriggerExternalSyncResponse, err error) {
	uid := l.ctx.Value("uid").(int64)

	result, err := l.svcCtx.RagRpc.TriggerExternalSync(l.ctx, &ragclient.TriggerExternalSyncReq{
		KbId: req.KbID,
		Uid:  uid,
	})
	if err != nil {
		return nil, err
	}

	return &types.TriggerExternalSyncResponse{
		SyncID: result.SyncId,
		Status: result.Status,
	}, nil
}
