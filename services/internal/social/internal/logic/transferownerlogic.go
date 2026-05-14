package logic

import (
	"context"
	"errors"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransferOwnerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferOwnerLogic {
	return &TransferOwnerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TransferOwnerLogic) TransferOwner(in *social.TransferOwnerReq) (*social.TransferOwnerResp, error) {
	g, err := l.svcCtx.SocialDao.GetGroupById(l.ctx, in.GroupId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if g.Owner != in.OldOwner {
		return nil, errors.New("只有群主可以转让")
	}

	newOwner, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.NewOwner)
	if err != nil {
		l.Logger.Error(err)
		return nil, errors.New("新群主不是群成员")
	}

	err = l.svcCtx.SocialDao.UpdateGroupOwner(l.ctx, in.GroupId, in.NewOwner)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.SocialDao.UpdateGroupMemberRole(l.ctx, in.GroupId, in.NewOwner, "OWNER")
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.SocialDao.UpdateGroupMemberRole(l.ctx, in.GroupId, in.OldOwner, "MEMBER")
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if newOwner.Role != "OWNER" {
		err = l.svcCtx.SocialDao.UpdateGroupMemberRole(l.ctx, in.GroupId, in.NewOwner, "OWNER")
		if err != nil {
			l.Logger.Error(err)
		}
	}

	err = l.svcCtx.RedisDao.DelGroupInfoCache(l.ctx, in.GroupId)
	if err != nil {
		l.Logger.Error("del group info cache error:", err)
	}

	return &social.TransferOwnerResp{}, nil
}
