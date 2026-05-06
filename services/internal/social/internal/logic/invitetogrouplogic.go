package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type InviteToGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInviteToGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteToGroupLogic {
	return &InviteToGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 邀请进群
func (l *InviteToGroupLogic) InviteToGroup(in *social.InviteToGroupReq) (*social.InviteToGroupResp, error) {
	// todo: add your logic here and delete this line

	return &social.InviteToGroupResp{}, nil
}
