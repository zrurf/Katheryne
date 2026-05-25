package logic

import (
	"context"
	"fmt"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSyncStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSyncStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSyncStatusLogic {
	return &GetSyncStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSyncStatusLogic) GetSyncStatus(in *rag.GetSyncStatusReq) (*rag.GetSyncStatusResp, error) {
	row, err := l.svcCtx.Storage.GetExternalSync(l.ctx, in.SyncId)
	if err != nil {
		return nil, fmt.Errorf("sync not found: %w", err)
	}

	lastSyncedAt := int64(0)
	if row.LastSyncedAt != nil {
		lastSyncedAt = row.LastSyncedAt.Unix()
	}

	return &rag.GetSyncStatusResp{
		Sync: &rag.ExternalSyncInfo{
			SyncId:       row.SyncID,
			KbId:         row.KbID,
			SourceType:   row.SourceType,
			SourceConfig: row.SourceConfig,
			SyncStatus:   row.SyncStatus,
			SyncError:    row.SyncError,
			LastSyncedAt: lastSyncedAt,
			CreatedAt:    row.CreatedAt.Unix(),
		},
	}, nil
}