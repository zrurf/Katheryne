package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeKBAuthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeKBAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeKBAuthLogic {
	return &RevokeKBAuthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeKBAuthLogic) RevokeKBAuth(req *types.RevokeKBAuthRequest) (resp *types.EmptyReponse, err error) {
	_, err = l.svcCtx.RagRpc.RevokeKBAuth(l.ctx, &ragclient.RevokeKBAuthReq{
		KbId:   req.KbID,
		BotId:  req.BotID,
		ConvId: req.ConvID,
	})
	if err != nil {
		return nil, err
	}
	return &types.EmptyReponse{}, nil
}