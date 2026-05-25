// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package rag

import (
	"context"
	"rag/ragclient"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSyncStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSyncStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSyncStatusLogic {
	return &GetSyncStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSyncStatusLogic) GetSyncStatus(req *types.GetSyncStatusRequest) (resp *types.GetSyncStatusResponse, err error) {
	result, err := l.svcCtx.RagRpc.GetSyncStatus(l.ctx, &ragclient.GetSyncStatusReq{
		SyncId: req.SyncID,
		KbId:   req.KbID,
	})
	if err != nil {
		return nil, err
	}

	sync := result.Sync
	return &types.GetSyncStatusResponse{
		Sync: types.ExternalSyncInfo{
			SyncID:       sync.SyncId,
			KbID:         sync.KbId,
			SourceType:   sync.SourceType,
			SourceConfig: sync.SourceConfig,
			SyncStatus:   sync.SyncStatus,
			SyncError:    sync.SyncError,
			LastSyncedAt: sync.LastSyncedAt,
			CreatedAt:    sync.CreatedAt,
		},
	}, nil
}
