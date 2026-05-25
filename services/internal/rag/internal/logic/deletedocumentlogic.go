package logic

import (
	"context"
	"fmt"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDocumentLogic {
	return &DeleteDocumentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDocumentLogic) DeleteDocument(in *rag.DeleteDocReq) (*rag.DeleteDocResp, error) {
	doc, err := l.svcCtx.Storage.GetDocument(l.ctx, in.DocId)
	if err != nil {
		return nil, fmt.Errorf("doc not found: %w", err)
	}

	// Soft-delete in PostgreSQL
	if err := l.svcCtx.Storage.DeleteDocument(l.ctx, in.DocId); err != nil {
		return nil, err
	}

	// Delete from Qdrant
	if l.svcCtx.Qdrant != nil {
		// Delete chunks by prefix
		chunkRows, _, err := l.svcCtx.Storage.GetDocChunks(l.ctx, in.DocId, 1, 10000)
		if err == nil {
			ids := make([]string, len(chunkRows))
			for i, c := range chunkRows {
				ids[i] = c.ChunkID
			}
			if len(ids) > 0 {
				_ = l.svcCtx.Qdrant.DeletePoints(l.ctx, doc.KbID, ids)
			}
		}
	}

	// Update KB doc count
	l.svcCtx.Storage.IncrementKBDocCount(l.ctx, doc.KbID, -1, -doc.FileSize)

	return &rag.DeleteDocResp{}, nil
}