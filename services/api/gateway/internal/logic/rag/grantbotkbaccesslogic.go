package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrantBotKBAccessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGrantBotKBAccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantBotKBAccessLogic {
	return &GrantBotKBAccessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GrantBotKBAccessLogic) GrantBotKBAccess(req *types.GrantBotKBAccessRequest) (resp *types.EmptyReponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	_, err = l.svcCtx.RagRpc.GrantBotKBAccess(l.ctx, &ragclient.GrantBotKBAccessReq{
		Uid:    uid,
		BotId:  req.BotID,
		ConvId: req.ConvID,
	})
	if err != nil {
		return nil, err
	}
	return &types.EmptyReponse{}, nil
}