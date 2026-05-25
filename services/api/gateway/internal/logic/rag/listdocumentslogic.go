package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDocumentsLogic {
	return &ListDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDocumentsLogic) ListDocuments(req *types.ListDocsRequest) (resp *types.ListDocsResponse, err error) {
	result, err := l.svcCtx.RagRpc.ListDocuments(l.ctx, &ragclient.ListDocsReq{
		KbId: req.KbID,
		Page: int32(req.Page),
		Size: int32(req.Size),
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.DocItem, 0, len(result.List))
	for _, doc := range result.List {
		list = append(list, types.DocItem{
			DocID:       doc.DocId,
			KbID:        doc.KbId,
			FileName:    doc.FileName,
			ContentType: doc.ContentType,
			FileSize:    doc.FileSize,
			Status:      doc.Status,
			ChunkCount:  doc.ChunkCount,
			ErrorMsg:    doc.ErrorMsg,
			CreatedAt:   doc.CreatedAt,
			UpdatedAt:   doc.UpdatedAt,
		})
	}
	return &types.ListDocsResponse{List: list, Total: result.Total}, nil
}