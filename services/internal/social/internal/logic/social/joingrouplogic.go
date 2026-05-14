package sociallogic

import (
	"context"
	"errors"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type JoinGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJoinGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinGroupLogic {
	return &JoinGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *JoinGroupLogic) JoinGroup(in *social.JoinGroupReq) (*social.JoinGroupResp, error) {
	g, err := l.svcCtx.SocialDao.GetGroupById(l.ctx, in.GroupId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if g.Status != "ACTIVE" {
		return nil, errors.New("群组不可用")
	}

	member, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.Uid)
	if err == nil && member != nil {
		return nil, errors.New("已经是群成员")
	}

	if g.VerifyMode == "ADMIN_CONFIRM" {
		req, err := l.svcCtx.SocialDao.InsertGroupJoinRequest(l.ctx, in.GroupId, in.Uid, in.Message)
		if err != nil {
			l.Logger.Error(err)
			return nil, err
		}
		return &social.JoinGroupResp{NeedConfirm: true, ReqId: req.Id}, nil
	}

	err = l.svcCtx.SocialDao.AddGroupMember(l.ctx, in.GroupId, in.Uid, "MEMBER", "", 0)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.SocialDao.IncrGroupMemberCount(l.ctx, in.GroupId, 1)
	if err != nil {
		l.Logger.Error(err)
	}

	convId, err := l.svcCtx.SocialDao.GetConversationByGroupId(l.ctx, in.GroupId)
	if err == nil {
		err = l.svcCtx.SocialDao.AddConvMember(l.ctx, convId, in.Uid)
		if err != nil {
			l.Logger.Error(err)
		}
	}

	err = l.svcCtx.RedisDao.DelGroupMembersCache(l.ctx, in.GroupId)
	if err != nil {
		l.Logger.Error("del group members cache error:", err)
	}

	return &social.JoinGroupResp{NeedConfirm: false}, nil
}