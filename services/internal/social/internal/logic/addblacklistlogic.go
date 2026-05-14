package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBlacklistLogic {
	return &AddBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddBlacklistLogic) AddBlacklist(in *social.AddBlacklistReq) (*social.AddBlacklistResp, error) {
	err := l.svcCtx.SocialDao.AddBlacklist(l.ctx, in.Uid, in.PeerUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelBlacklistCache(l.ctx, in.Uid)
	if err != nil {
		l.Logger.Error("del blacklist cache error:", err)
	}

	return &social.AddBlacklistResp{}, nil
}
