package sociallogic

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

func (l *InviteToGroupLogic) InviteToGroup(in *social.InviteToGroupReq) (*social.InviteToGroupResp, error) {
	isMember, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.InviterUid)
	if err != nil || isMember == nil {
		return &social.InviteToGroupResp{}, nil
	}

	var failedUids []int64
	for _, inviteeUid := range in.InviteeUids {
		_, err = l.svcCtx.SocialDao.InsertGroupInvite(l.ctx, in.GroupId, in.InviterUid, inviteeUid, in.Message)
		if err != nil {
			failedUids = append(failedUids, inviteeUid)
		}
	}

	return &social.InviteToGroupResp{FailedUids: failedUids}, nil
}