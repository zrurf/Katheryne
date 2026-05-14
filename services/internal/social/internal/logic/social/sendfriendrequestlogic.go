package sociallogic

import (
	"context"
	"errors"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendFriendRequestLogic {
	return &SendFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 好友
func (l *SendFriendRequestLogic) SendFriendRequest(in *social.SendFriendReq) (*social.SendFriendResp, error) {
	if in.Uid == in.ToUid {
		return nil, errors.New("不能添加自己为好友")
	}

	isFriend, err := l.svcCtx.SocialDao.IsFriend(l.ctx, in.Uid, in.ToUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	if isFriend {
		return nil, errors.New("已经是好友")
	}

	req, err := l.svcCtx.SocialDao.InsertFriendRequest(l.ctx, in.Uid, in.ToUid, in.Message)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &social.SendFriendResp{ReqId: req.Id}, nil
}
