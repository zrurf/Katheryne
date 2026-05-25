package logic

import (
	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
)

type AuthorizeKBLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthorizeKBLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthorizeKBLogic {
	return &AuthorizeKBLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthorizeKBLogic) AuthorizeKB(in *rag.AuthorizeKBReq) (*rag.AuthorizeKBResp, error) {
	if err := l.svcCtx.Storage.AuthorizeKB(l.ctx, in.KbId, in.BotId, in.ConvId, in.Permission); err != nil {
		return nil, err
	}
	return &rag.AuthorizeKBResp{}, nil
}