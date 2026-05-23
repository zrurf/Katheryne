package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupLogic {
	return &UpdateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateGroupLogic) UpdateGroup(in *social.UpdateGroupReq) (*social.UpdateGroupResp, error) {
	err := l.svcCtx.SocialDao.UpdateGroup(l.ctx, in.GroupId, in.Name, in.Avatar, in.VerifyMode)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelGroupInfoCache(l.ctx, in.GroupId)
	if err != nil {
		l.Logger.Error("del group info cache error:", err)
	}

	// Sync conversation table so group avatar shows in chat list
	err = l.svcCtx.SocialDao.UpdateConversationByGroupId(l.ctx, in.GroupId, in.Name, in.Avatar)
	if err != nil {
		l.Logger.Error("update conversation error:", err)
	} else {
		members, err := l.svcCtx.SocialDao.ListGroupMembers(l.ctx, in.GroupId, "")
		if err != nil {
			l.Logger.Error("list group members error:", err)
		} else {
			for _, m := range members {
				_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, m.Uid)
			}
		}
	}

	return &social.UpdateGroupResp{}, nil
}
