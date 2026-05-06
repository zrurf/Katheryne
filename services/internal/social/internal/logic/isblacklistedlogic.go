package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsBlacklistedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsBlacklistedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsBlacklistedLogic {
	return &IsBlacklistedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsBlacklistedLogic) IsBlacklisted(in *social.IsBlacklistedReq) (*social.IsBlacklistedResp, error) {
	// todo: add your logic here and delete this line

	return &social.IsBlacklistedResp{}, nil
}
