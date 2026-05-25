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

type ListExternalSyncsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListExternalSyncsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListExternalSyncsLogic {
	return &ListExternalSyncsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListExternalSyncsLogic) ListExternalSyncs(req *types.ListExternalSyncsRequest) (resp *types.ListExternalSyncsResponse, err error) {
	uid := l.ctx.Value("uid").(int64)

	result, err := l.svcCtx.RagRpc.ListExternalSyncs(l.ctx, &ragclient.ListExternalSyncsReq{
		KbId: req.KbID,
		Uid:  uid,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ExternalSyncInfo, 0, len(result.List))
	for _, s := range result.List {
		list = append(list, types.ExternalSyncInfo{
			SyncID:       s.SyncId,
			KbID:         s.KbId,
			SourceType:   s.SourceType,
			SourceConfig: s.SourceConfig,
			SyncStatus:   s.SyncStatus,
			SyncError:    s.SyncError,
			LastSyncedAt: s.LastSyncedAt,
			CreatedAt:    s.CreatedAt,
		})
	}

	return &types.ListExternalSyncsResponse{List: list}, nil
}
