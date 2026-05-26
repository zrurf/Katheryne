package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteKnowledgeBaseLogic {
	return &DeleteKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteKnowledgeBaseLogic) DeleteKnowledgeBase(req *types.DeleteKBRequest) (resp *types.EmptyReponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	_, err = l.svcCtx.RagRpc.DeleteKnowledgeBase(l.ctx, &ragclient.DeleteKBReq{
		Uid:  uid,
		KbId: req.KbID,
	})
	if err != nil {
		return nil, err
	}
	return &types.EmptyReponse{}, nil
}
