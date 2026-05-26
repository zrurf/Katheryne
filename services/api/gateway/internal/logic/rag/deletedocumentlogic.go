package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDocumentLogic {
	return &DeleteDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDocumentLogic) DeleteDocument(req *types.DeleteDocRequest) (resp *types.EmptyReponse, err error) {
	_, err = l.svcCtx.RagRpc.DeleteDocument(l.ctx, &ragclient.DeleteDocReq{
		KbId:  req.KbID,
		DocId: req.DocID,
	})
	if err != nil {
		return nil, err
	}
	return &types.EmptyReponse{}, nil
}
