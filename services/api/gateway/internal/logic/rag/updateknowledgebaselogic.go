package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateKnowledgeBaseLogic {
	return &UpdateKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateKnowledgeBaseLogic) UpdateKnowledgeBase(req *types.UpdateKBRequest) (resp *types.EmptyReponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	_, err = l.svcCtx.RagRpc.UpdateKnowledgeBase(l.ctx, &ragclient.UpdateKBReq{
		Uid:         uid,
		KbId:        req.KbID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &types.EmptyReponse{}, nil
}