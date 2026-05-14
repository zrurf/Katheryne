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
	ok, err := l.svcCtx.SocialDao.IsBlacklisted(l.ctx, in.Uid, in.PeerUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &social.IsBlacklistedResp{Blacklisted: ok}, nil
}
