package logic

import (
	"context"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDocumentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDocumentsLogic {
	return &ListDocumentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListDocumentsLogic) ListDocuments(in *rag.ListDocsReq) (*rag.ListDocsResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	size := in.Size
	if size < 1 || size > 100 {
		size = 20
	}

	rows, total, err := l.svcCtx.Storage.ListDocuments(l.ctx, in.KbId, page, size)
	if err != nil {
		return nil, err
	}

	list := make([]*rag.Document, 0, len(rows))
	for _, r := range rows {
		list = append(list, &rag.Document{
			DocId:       r.DocID,
			KbId:        r.KbID,
			FileName:    r.FileName,
			ContentType: r.ContentType,
			FileSize:    r.FileSize,
			OssIndex:    r.OssIndex,
			Status:      r.Status,
			ChunkCount:  r.ChunkCount,
			ErrorMsg:    r.ErrorMsg,
			CreatedAt:   r.CreatedAt.UnixMilli(),
			UpdatedAt:   r.UpdatedAt.UnixMilli(),
		})
	}

	return &rag.ListDocsResp{List: list, Total: total}, nil
}