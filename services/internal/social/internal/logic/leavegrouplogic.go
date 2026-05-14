package logic

import (
	"context"
	"errors"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type LeaveGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLeaveGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveGroupLogic {
	return &LeaveGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LeaveGroupLogic) LeaveGroup(in *social.LeaveGroupReq) (*social.LeaveGroupResp, error) {
	member, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, errors.New("不是群成员")
	}

	if member.Role == "OWNER" {
		return nil, errors.New("群主不能退群，请先转让群主")
	}

	err = l.svcCtx.SocialDao.RemoveGroupMember(l.ctx, in.GroupId, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.SocialDao.IncrGroupMemberCount(l.ctx, in.GroupId, -1)
	if err != nil {
		l.Logger.Error(err)
	}

	err = l.svcCtx.RedisDao.DelGroupMembersCache(l.ctx, in.GroupId)
	if err != nil {
		l.Logger.Error("del group members cache error:", err)
	}

	return &social.LeaveGroupResp{}, nil
}
