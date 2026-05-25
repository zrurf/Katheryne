package logic

import (
	"context"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListKnowledgeBasesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListKnowledgeBasesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKnowledgeBasesLogic {
	return &ListKnowledgeBasesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListKnowledgeBasesLogic) ListKnowledgeBases(in *rag.ListKBsReq) (*rag.ListKBsResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	size := in.Size
	if size < 1 || size > 100 {
		size = 20
	}

	rows, total, err := l.svcCtx.Storage.ListKBs(l.ctx, in.OwnerUid, page, size)
	if err != nil {
		return nil, err
	}

	list := make([]*rag.KnowledgeBase, 0, len(rows))
	for _, r := range rows {
		list = append(list, &rag.KnowledgeBase{
			KbId:        r.KbID,
			Name:        r.Name,
			Description: r.Description,
			OwnerUid:    r.OwnerUID,
			Status:      r.Status,
			DocCount:    r.DocCount,
			ChunkCount:  r.ChunkCount,
			TotalSize:   r.TotalSize,
			CreatedAt:   r.CreatedAt.UnixMilli(),
			UpdatedAt:   r.UpdatedAt.UnixMilli(),
		})
	}

	return &rag.ListKBsResp{List: list, Total: total}, nil
}