package logic

import (
	"context"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteKnowledgeBaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteKnowledgeBaseLogic {
	return &DeleteKnowledgeBaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteKnowledgeBaseLogic) DeleteKnowledgeBase(in *rag.DeleteKBReq) (*rag.DeleteKBResp, error) {
	row, err := l.svcCtx.Storage.GetKB(l.ctx, in.KbId)
	if err != nil {
		return nil, errKBNotFound
	}
	if row.OwnerUID != in.Uid {
		return nil, errPermissionDenied
	}

	// Soft-delete in PostgreSQL
	if err := l.svcCtx.Storage.DeleteKB(l.ctx, in.KbId); err != nil {
		return nil, err
	}

	// Drop Qdrant collection
	if l.svcCtx.Qdrant != nil {
		if err := l.svcCtx.Qdrant.DeleteCollection(l.ctx, in.KbId); err != nil {
			l.Errorf("delete qdrant collection %s: %v", in.KbId, err)
		}
	}

	// Remove HugeGraph data
	if l.svcCtx.HugeGraph != nil {
		if err := l.svcCtx.HugeGraph.DeleteGraphByKB(l.ctx, in.KbId); err != nil {
			l.Errorf("delete hugegraph data %s: %v", in.KbId, err)
		}
	}

	return &rag.DeleteKBResp{}, nil
}