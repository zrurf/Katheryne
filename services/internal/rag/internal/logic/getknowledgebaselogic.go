package logic

import (
	"context"
	"fmt"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeBaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeBaseLogic {
	return &GetKnowledgeBaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetKnowledgeBaseLogic) GetKnowledgeBase(in *rag.GetKBReq) (*rag.GetKBResp, error) {
	row, err := l.svcCtx.Storage.GetKB(l.ctx, in.KbId)
	if err != nil {
		return nil, fmt.Errorf("kb not found: %w", err)
	}

	if row.OwnerUID != in.Uid {
		return nil, fmt.Errorf("permission denied")
	}

	return &rag.GetKBResp{
		Kb: &rag.KnowledgeBase{
			KbId:        row.KbID,
			Name:        row.Name,
			Description: row.Description,
			OwnerUid:    row.OwnerUID,
			Status:      row.Status,
			DocCount:    row.DocCount,
			ChunkCount:  row.ChunkCount,
			TotalSize:   row.TotalSize,
			CreatedAt:   row.CreatedAt.UnixMilli(),
			UpdatedAt:   row.UpdatedAt.UnixMilli(),
		},
	}, nil
}