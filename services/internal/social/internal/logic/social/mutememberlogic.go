package sociallogic

import (
	"context"
	"errors"
	"time"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type MuteMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMuteMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MuteMemberLogic {
	return &MuteMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MuteMemberLogic) MuteMember(in *social.MuteMemberReq) (*social.MuteMemberResp, error) {
	operator, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.OperatorUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, errors.New("操作者不是群成员")
	}

	if operator.Role != "OWNER" && operator.Role != "ADMIN" {
		return nil, errors.New("无权禁言成员")
	}

	target, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.TargetUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, errors.New("目标用户不是群成员")
	}

	if target.Role == "OWNER" {
		return nil, errors.New("不能禁言群主")
	}

	var muteUntil *time.Time
	if in.MuteUntil > 0 {
		t := time.Unix(in.MuteUntil, 0)
		muteUntil = &t
	}

	err = l.svcCtx.SocialDao.UpdateMemberMute(l.ctx, in.GroupId, in.TargetUid, muteUntil)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &social.MuteMemberResp{}, nil
}