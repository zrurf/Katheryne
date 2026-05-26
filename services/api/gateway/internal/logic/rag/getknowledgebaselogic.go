package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeBaseLogic {
	return &GetKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeBaseLogic) GetKnowledgeBase(req *types.GetKBRequest) (resp *types.GetKBResponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.RagRpc.GetKnowledgeBase(l.ctx, &ragclient.GetKBReq{
		KbId: req.KbID,
		Uid:  uid,
	})
	if err != nil {
		return nil, err
	}
	kb := result.Kb
	return &types.GetKBResponse{KB: types.KBItem{
		KbID:        kb.KbId,
		Name:        kb.Name,
		Description: kb.Description,
		Status:      kb.Status,
		DocCount:    kb.DocCount,
		ChunkCount:  kb.ChunkCount,
		TotalSize:   kb.TotalSize,
		CreatedAt:   kb.CreatedAt,
		UpdatedAt:   kb.UpdatedAt,
	}}, nil
}
