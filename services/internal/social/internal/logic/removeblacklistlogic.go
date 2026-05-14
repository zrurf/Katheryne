package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveBlacklistLogic {
	return &RemoveBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveBlacklistLogic) RemoveBlacklist(in *social.RemoveBlacklistReq) (*social.RemoveBlacklistResp, error) {
	err := l.svcCtx.SocialDao.RemoveBlacklist(l.ctx, in.Uid, in.PeerUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelBlacklistCache(l.ctx, in.Uid)
	if err != nil {
		l.Logger.Error("del blacklist cache error:", err)
	}

	return &social.RemoveBlacklistResp{}, nil
}
