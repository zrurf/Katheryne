package logic

import (
	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
)

type ListBotKnowledgeBasesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBotKnowledgeBasesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBotKnowledgeBasesLogic {
	return &ListBotKnowledgeBasesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListBotKnowledgeBasesLogic) ListBotKnowledgeBases(in *rag.ListBotKBsReq) (*rag.ListBotKBsResp, error) {
	rows, err := l.svcCtx.Storage.ListBotKBs(l.ctx, in.BotId, in.ConvId, in.Uid)
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

	return &rag.ListBotKBsResp{List: list}, nil
}