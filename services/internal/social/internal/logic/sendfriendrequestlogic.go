package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &social.SendFriendResp{}, nil
}
