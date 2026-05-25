package logic

import (
	"context"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateKnowledgeBaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateKnowledgeBaseLogic {
	return &UpdateKnowledgeBaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateKnowledgeBaseLogic) UpdateKnowledgeBase(in *rag.UpdateKBReq) (*rag.UpdateKBResp, error) {
	row, err := l.svcCtx.Storage.GetKB(l.ctx, in.KbId)
	if err != nil {
		return nil, err
	}
	if row.OwnerUID != in.Uid {
		return nil, errPermissionDenied
	}

	if err := l.svcCtx.Storage.UpdateKB(l.ctx, in.KbId, in.Name, in.Description); err != nil {
		return nil, err
	}
	return &rag.UpdateKBResp{}, nil
}