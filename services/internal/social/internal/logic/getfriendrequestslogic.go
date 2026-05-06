package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestsLogic {
	return &GetFriendRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFriendRequestsLogic) GetFriendRequests(in *social.GetFriendRequestsReq) (*social.GetFriendRequestsResp, error) {
	// todo: add your logic here and delete this line

	return &social.GetFriendRequestsResp{}, nil
}
