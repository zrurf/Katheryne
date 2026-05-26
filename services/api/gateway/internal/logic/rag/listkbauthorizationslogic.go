package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListKBAuthorizationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListKBAuthorizationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKBAuthorizationsLogic {
	return &ListKBAuthorizationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListKBAuthorizationsLogic) ListKBAuthorizations(req *types.ListKBAuthsRequest) (resp *types.ListKBAuthsResponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.RagRpc.ListKBAuthorizations(l.ctx, &ragclient.ListKBAuthsReq{
		Uid:   uid,
		KbId:  req.KbID,
		BotId: req.BotID,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.KBAuthItem, 0, len(result.List))
	for _, a := range result.List {
		list = append(list, types.KBAuthItem{
			KbID:       a.KbId,
			KbName:     a.KbName,
			BotID:      a.BotId,
			ConvID:     a.ConvId,
			Permission: a.Permission,
			GrantedAt:  a.GrantedAt,
		})
	}
	return &types.ListKBAuthsResponse{List: list}, nil
}
