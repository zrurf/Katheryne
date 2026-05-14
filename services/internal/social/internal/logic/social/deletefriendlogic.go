package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteFriendLogic) DeleteFriend(in *social.DeleteFriendReq) (*social.DeleteFriendResp, error) {
	err := l.svcCtx.SocialDao.DeleteFriendship(l.ctx, in.Uid, in.PeerUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelFriendshipCache(l.ctx, in.Uid)
	if err != nil {
		l.Logger.Error("del friendship cache error:", err)
	}
	err = l.svcCtx.RedisDao.DelFriendshipCache(l.ctx, in.PeerUid)
	if err != nil {
		l.Logger.Error("del friendship cache error:", err)
	}

	return &social.DeleteFriendResp{}, nil
}