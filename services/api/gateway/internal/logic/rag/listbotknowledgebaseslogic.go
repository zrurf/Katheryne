package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBotKnowledgeBasesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBotKnowledgeBasesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBotKnowledgeBasesLogic {
	return &ListBotKnowledgeBasesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListBotKnowledgeBasesLogic) ListBotKnowledgeBases(req *types.ListBotKBsRequest) (resp *types.ListKBsResponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.RagRpc.ListBotKnowledgeBases(l.ctx, &ragclient.ListBotKBsReq{
		BotId:  req.BotID,
		ConvId: req.ConvID,
		Uid:    uid,
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
	return &types.ListKBsResponse{List: list, Total: int64(len(list))}, nil
}
