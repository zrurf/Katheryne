package logic

import (
	"context"
	"fmt"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDocumentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDocumentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDocumentStatusLogic {
	return &GetDocumentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDocumentStatusLogic) GetDocumentStatus(in *rag.GetDocStatusReq) (*rag.GetDocStatusResp, error) {
	row, err := l.svcCtx.Storage.GetDocument(l.ctx, in.DocId)
	if err != nil {
		return nil, fmt.Errorf("doc not found: %w", err)
	}

	return &rag.GetDocStatusResp{
		Doc: &rag.Document{
			DocId:       row.DocID,
			KbId:        row.KbID,
			FileName:    row.FileName,
			ContentType: row.ContentType,
			FileSize:    row.FileSize,
			OssIndex:    row.OssIndex,
			Status:      row.Status,
			ChunkCount:  row.ChunkCount,
			ErrorMsg:    row.ErrorMsg,
			CreatedAt:   row.CreatedAt.UnixMilli(),
			UpdatedAt:   row.UpdatedAt.UnixMilli(),
		},
	}, nil
}