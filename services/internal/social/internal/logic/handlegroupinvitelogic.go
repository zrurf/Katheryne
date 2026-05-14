package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupInviteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupInviteLogic {
	return &HandleGroupInviteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HandleGroupInviteLogic) HandleGroupInvite(in *social.HandleGroupInviteReq) (*social.HandleGroupInviteResp, error) {
	invite, err := l.svcCtx.SocialDao.GetGroupInviteById(l.ctx, in.InviteId)
	if err != nil {
		return nil, err
	}
	if invite.Invitee != in.Uid {
		return &social.HandleGroupInviteResp{}, nil
	}

	if in.Action == "accept" {
		err = l.svcCtx.SocialDao.UpdateGroupInviteStatus(l.ctx, in.InviteId, "accepted")
		if err != nil {
			return nil, err
		}
		err = l.svcCtx.SocialDao.AddGroupMember(l.ctx, invite.GroupId, in.Uid, "MEMBER", "", invite.Inviter)
		if err != nil {
			return nil, err
		}
		err = l.svcCtx.SocialDao.IncrGroupMemberCount(l.ctx, invite.GroupId, 1)
		if err != nil {
			return nil, err
		}
	} else {
		err = l.svcCtx.SocialDao.UpdateGroupInviteStatus(l.ctx, in.InviteId, "rejected")
		if err != nil {
			return nil, err
		}
	}

	return &social.HandleGroupInviteResp{}, nil
}
