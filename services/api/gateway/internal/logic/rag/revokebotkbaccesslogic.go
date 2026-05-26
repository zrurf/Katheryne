package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeBotKBAccessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeBotKBAccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeBotKBAccessLogic {
	return &RevokeBotKBAccessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeBotKBAccessLogic) RevokeBotKBAccess(req *types.RevokeBotKBAccessRequest) (resp *types.EmptyReponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	_, err = l.svcCtx.RagRpc.RevokeBotKBAccess(l.ctx, &ragclient.RevokeBotKBAccessReq{
		Uid:    uid,
		BotId:  req.BotID,
		ConvId: req.ConvID,
	})
	if err != nil {
		return nil, err
	}
	return &types.EmptyReponse{}, nil
}
