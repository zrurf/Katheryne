package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsFriendLogic {
	return &IsFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsFriendLogic) IsFriend(in *social.IsFriendReq) (*social.IsFriendResp, error) {
	ok, err := l.svcCtx.SocialDao.IsFriend(l.ctx, in.Uid, in.PeerUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &social.IsFriendResp{IsFriend: ok}, nil
}