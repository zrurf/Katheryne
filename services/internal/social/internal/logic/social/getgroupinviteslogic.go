package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInvitesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupInvitesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInvitesLogic {
	return &GetGroupInvitesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupInvitesLogic) GetGroupInvites(in *social.GetGroupInvitesReq) (*social.GetGroupInvitesResp, error) {
	invites, err := l.svcCtx.SocialDao.ListGroupInvitesByInvitee(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}

	list := make([]*social.GroupInviteItem, len(invites))
	for i, inv := range invites {
		group, _ := l.svcCtx.SocialDao.GetGroupById(l.ctx, inv.GroupId)
		groupName := ""
		groupAvatar := ""
		if group != nil {
			groupName = group.Name
			if group.Avatar.Valid {
				groupAvatar = group.Avatar.String
			}
		}
		inviter, _ := l.svcCtx.UserDBDao.GetUserById(l.ctx, inv.Inviter)
		inviterName := ""
		if inviter != nil {
			inviterName = inviter.Name
		}
		list[i] = &social.GroupInviteItem{
			Id:          inv.Id,
			GroupId:     inv.GroupId,
			GroupName:   groupName,
			GroupAvatar: groupAvatar,
			InviterUid:  inv.Inviter,
			InviterName: inviterName,
			CreatedAt:   inv.CreatedAt.Unix(),
		}
	}
	return &social.GetGroupInvitesResp{
		List: list,
	}, nil
}
