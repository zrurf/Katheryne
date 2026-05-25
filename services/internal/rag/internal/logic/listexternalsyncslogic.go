package logic

import (
	"context"
	"fmt"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListExternalSyncsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListExternalSyncsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListExternalSyncsLogic {
	return &ListExternalSyncsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListExternalSyncsLogic) ListExternalSyncs(in *rag.ListExternalSyncsReq) (*rag.ListExternalSyncsResp, error) {
	rows, err := l.svcCtx.Storage.ListExternalSyncs(l.ctx, in.KbId)
	if err != nil {
		return nil, fmt.Errorf("list syncs: %w", err)
	}

	list := make([]*rag.ExternalSyncInfo, 0, len(rows))
	for _, row := range rows {
		lastSyncedAt := int64(0)
		if row.LastSyncedAt != nil {
			lastSyncedAt = row.LastSyncedAt.Unix()
		}
		list = append(list, &rag.ExternalSyncInfo{
			SyncId:       row.SyncID,
			KbId:         row.KbID,
			SourceType:   row.SourceType,
			SourceConfig: row.SourceConfig,
			SyncStatus:   row.SyncStatus,
			SyncError:    row.SyncError,
			LastSyncedAt: lastSyncedAt,
			CreatedAt:    row.CreatedAt.Unix(),
		})
	}

	return &rag.ListExternalSyncsResp{List: list}, nil
}