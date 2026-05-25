package logic

import (
	"context"
	"encoding/json"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDocumentChunksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDocumentChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDocumentChunksLogic {
	return &GetDocumentChunksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDocumentChunksLogic) GetDocumentChunks(in *rag.GetDocChunksReq) (*rag.GetDocChunksResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	size := in.Size
	if size < 1 || size > 100 {
		size = 20
	}

	rows, total, err := l.svcCtx.Storage.GetDocChunks(l.ctx, in.DocId, page, size)
	if err != nil {
		return nil, err
	}

	list := make([]*rag.ChunkInfo, 0, len(rows))
	for _, r := range rows {
		var entities []string
		if r.Entities != "" {
			_ = json.Unmarshal([]byte(r.Entities), &entities)
		}
		if entities == nil {
			entities = []string{}
		}
		list = append(list, &rag.ChunkInfo{
			ChunkId:    r.ChunkID,
			Content:    r.Content,
			ChunkIndex: r.ChunkIndex,
			TokenCount: r.TokenCount,
			Entities:   entities,
		})
	}

	return &rag.GetDocChunksResp{List: list, Total: total}, nil
}