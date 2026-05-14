package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserGroupsLogic {
	return &GetUserGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserGroupsLogic) GetUserGroups(in *social.GetUserGroupsReq) (*social.GetUserGroupsResp, error) {
	groups, err := l.svcCtx.SocialDao.ListUserGroups(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}

	list := make([]*social.GroupInfo, len(groups))
	for i, g := range groups {
		list[i] = &social.GroupInfo{
			GroupId:     g.GroupId,
			Name:        g.Name,
			Avatar:      g.Avatar.String,
			Owner:       g.Owner,
			MemberCount: g.MemberCount,
			Status:      g.Status,
			VerifyMode:  g.VerifyMode,
			CreatedAt:   g.CreatedAt.Unix(),
		}
	}
	return &social.GetUserGroupsResp{
		List: list,
	}, nil
}
