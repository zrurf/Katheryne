package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthorizeKBLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthorizeKBLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthorizeKBLogic {
	return &AuthorizeKBLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthorizeKBLogic) AuthorizeKB(req *types.AuthorizeKBRequest) (resp *types.EmptyReponse, err error) {
	_, err = l.svcCtx.RagRpc.AuthorizeKB(l.ctx, &ragclient.AuthorizeKBReq{
		KbId:       req.KbID,
		BotId:      req.BotID,
		ConvId:     req.ConvID,
		Permission: req.Permission,
	})
	if err != nil {
		return nil, err
	}
	return &types.EmptyReponse{}, nil
}
