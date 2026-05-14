package logic

import (
	"context"
	"time"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsGroupMutedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsGroupMutedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsGroupMutedLogic {
	return &IsGroupMutedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsGroupMutedLogic) IsGroupMuted(in *social.IsGroupMutedReq) (*social.IsGroupMutedResp, error) {
	member, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.Uid)
	if err != nil || member == nil {
		return &social.IsGroupMutedResp{Muted: false}, nil
	}

	isMuted := false
	var muteUntil int64
	if member.MuteUntil.Valid && member.MuteUntil.Time.After(time.Now()) {
		isMuted = true
		muteUntil = member.MuteUntil.Time.Unix()
	}
	return &social.IsGroupMutedResp{Muted: isMuted, MuteUntil: muteUntil}, nil
}
