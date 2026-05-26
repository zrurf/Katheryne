package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDocumentChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDocumentChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDocumentChunksLogic {
	return &GetDocumentChunksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDocumentChunksLogic) GetDocumentChunks(req *types.GetDocChunksRequest) (resp *types.GetDocChunksResponse, err error) {
	result, err := l.svcCtx.RagRpc.GetDocumentChunks(l.ctx, &ragclient.GetDocChunksReq{
		KbId:  req.KbID,
		DocId: req.DocID,
		Page:  int32(req.Page),
		Size:  int32(req.Size),
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.ChunkItem, 0, len(result.List))
	for _, c := range result.List {
		list = append(list, types.ChunkItem{
			ChunkID:    c.ChunkId,
			Content:    c.Content,
			ChunkIndex: c.ChunkIndex,
			TokenCount: c.TokenCount,
			Entities:   c.Entities,
		})
	}
	return &types.GetDocChunksResponse{List: list, Total: result.Total}, nil
}
