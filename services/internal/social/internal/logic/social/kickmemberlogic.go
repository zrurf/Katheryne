package sociallogic

import (
	"context"
	"errors"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type KickMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewKickMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KickMemberLogic {
	return &KickMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *KickMemberLogic) KickMember(in *social.KickMemberReq) (*social.KickMemberResp, error) {
	operator, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.OperatorUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, errors.New("操作者不是群成员")
	}

	if operator.Role != "OWNER" && operator.Role != "ADMIN" {
		return nil, errors.New("无权踢出成员")
	}

	target, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.TargetUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, errors.New("目标用户不是群成员")
	}

	if target.Role == "OWNER" {
		return nil, errors.New("不能踢出群主")
	}

	if operator.Role == "ADMIN" && target.Role == "ADMIN" {
		return nil, errors.New("管理员不能踢出管理员")
	}

	err = l.svcCtx.SocialDao.RemoveGroupMember(l.ctx, in.GroupId, in.TargetUid)
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

	return &social.KickMemberResp{}, nil
}