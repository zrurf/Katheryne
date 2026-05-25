package logic

import (
	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
)

type RevokeKBAuthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevokeKBAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeKBAuthLogic {
	return &RevokeKBAuthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RevokeKBAuthLogic) RevokeKBAuth(in *rag.RevokeKBAuthReq) (*rag.RevokeKBAuthResp, error) {
	if err := l.svcCtx.Storage.RevokeKBAuth(l.ctx, in.KbId, in.BotId, in.ConvId); err != nil {
		return nil, err
	}
	return &rag.RevokeKBAuthResp{}, nil
}