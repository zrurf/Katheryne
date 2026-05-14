package logic

import (
	"context"
	"time"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateGroupLogic) CreateGroup(in *social.CreateGroupReq) (*social.CreateGroupResp, error) {
	groupId := time.Now().UnixNano()

	g, err := l.svcCtx.SocialDao.InsertGroup(l.ctx, groupId, in.OwnerUid, in.Name, in.Avatar, in.VerifyMode)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.SocialDao.AddGroupMember(l.ctx, groupId, in.OwnerUid, "OWNER", "", 0)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	for _, uid := range in.MemberUids {
		if uid == in.OwnerUid {
			continue
		}
		err = l.svcCtx.SocialDao.AddGroupMember(l.ctx, groupId, uid, "MEMBER", "", in.OwnerUid)
		if err != nil {
			l.Logger.Error("add member error:", err)
			continue
		}
		err = l.svcCtx.SocialDao.IncrGroupMemberCount(l.ctx, groupId, 1)
		if err != nil {
			l.Logger.Error("incr member count error:", err)
		}
	}

	convId, err := l.svcCtx.SocialDao.InsertConversation(l.ctx, "GROUP", groupId, in.Name, in.Avatar)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.SocialDao.AddConvMember(l.ctx, convId, in.OwnerUid)
	if err != nil {
		l.Logger.Error(err)
	}
	for _, uid := range in.MemberUids {
		if uid == in.OwnerUid {
			continue
		}
		err = l.svcCtx.SocialDao.AddConvMember(l.ctx, convId, uid)
		if err != nil {
			l.Logger.Error("add conv member error:", err)
		}
	}

	return &social.CreateGroupResp{
		GroupId: g.GroupId,
		ConvId:  convId,
	}, nil
}
