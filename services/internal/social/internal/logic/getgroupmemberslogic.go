package logic

import (
	"context"
	"fmt"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMembersLogic {
	return &GetGroupMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupMembersLogic) GetGroupMembers(in *social.GetGroupMembersReq) (*social.GetGroupMembersResp, error) {
	list, err := l.svcCtx.SocialDao.ListGroupMembers(l.ctx, in.GroupId, in.Role)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	items := make([]*social.GroupMemberItem, len(list))
	for i, m := range list {
		user, err := l.svcCtx.UserDBDao.GetUserById(l.ctx, m.Uid)
		name := fmt.Sprintf("用户%d", m.Uid)
		avatar := ""
		if err == nil && user != nil {
			if user.Name != "" {
				name = user.Name
			}
			avatar = nullString(user.Avatar)
		}
		item := &social.GroupMemberItem{
			Uid:      m.Uid,
			Name:     name,
			Avatar:   avatar,
			Role:     m.Role,
			Nick:     nullString(m.Nick),
			JoinTime: m.JoinTime.UnixMilli(),
		}
		if m.MuteUntil.Valid {
			item.MuteUntil = m.MuteUntil.Time.Unix()
		}
		items[i] = item
	}

	return &social.GetGroupMembersResp{List: items}, nil
}
