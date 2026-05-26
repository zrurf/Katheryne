package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDocumentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDocumentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDocumentStatusLogic {
	return &GetDocumentStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDocumentStatusLogic) GetDocumentStatus(req *types.GetDocStatusRequest) (resp *types.GetDocStatusResponse, err error) {
	result, err := l.svcCtx.RagRpc.GetDocumentStatus(l.ctx, &ragclient.GetDocStatusReq{
		DocId: req.DocID,
		KbId:  req.KbID,
	})
	if err != nil {
		return nil, err
	}
	doc := result.Doc
	return &types.GetDocStatusResponse{Doc: types.DocItem{
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
	}}, nil
}
