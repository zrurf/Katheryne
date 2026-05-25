package logic

import (
	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
)

type ListKBAuthorizationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListKBAuthorizationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKBAuthorizationsLogic {
	return &ListKBAuthorizationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListKBAuthorizationsLogic) ListKBAuthorizations(in *rag.ListKBAuthsReq) (*rag.ListKBAuthsResp, error) {
	rows, err := l.svcCtx.Storage.ListKBAuths(l.ctx, in.Uid, in.KbId, in.BotId)
	if err != nil {
		return nil, err
	}

	list := make([]*rag.KBAuthorization, 0, len(rows))
	for _, r := range rows {
		list = append(list, &rag.KBAuthorization{
			KbId:       r.KbID,
			BotId:      r.BotID,
			ConvId:     r.ConvID,
			Permission: r.Permission,
			GrantedAt:  r.GrantedAt.UnixMilli(),
		})
	}

	return &rag.ListKBAuthsResp{List: list}, nil
}