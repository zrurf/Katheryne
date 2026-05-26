package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListKnowledgeBasesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListKnowledgeBasesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKnowledgeBasesLogic {
	return &ListKnowledgeBasesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListKnowledgeBasesLogic) ListKnowledgeBases(req *types.ListKBsRequest) (resp *types.ListKBsResponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.RagRpc.ListKnowledgeBases(l.ctx, &ragclient.ListKBsReq{
		OwnerUid: uid,
		Page:     int32(req.Page),
		Size:     int32(req.Size),
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.KBItem, 0, len(result.List))
	for _, kb := range result.List {
		list = append(list, types.KBItem{
			KbID:        kb.KbId,
			Name:        kb.Name,
			Description: kb.Description,
			Status:      kb.Status,
			DocCount:    kb.DocCount,
			ChunkCount:  kb.ChunkCount,
			TotalSize:   kb.TotalSize,
			CreatedAt:   kb.CreatedAt,
			UpdatedAt:   kb.UpdatedAt,
		})
	}
	return &types.ListKBsResponse{List: list, Total: result.Total}, nil
}
