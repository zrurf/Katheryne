package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetGroupAnnouncementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetGroupAnnouncementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetGroupAnnouncementLogic {
	return &SetGroupAnnouncementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetGroupAnnouncementLogic) SetGroupAnnouncement(in *social.SetGroupAnnouncementReq) (*social.SetGroupAnnouncementResp, error) {
	err := l.svcCtx.SocialDao.UpdateGroupAnnouncement(l.ctx, in.GroupId, in.Content)
	if err != nil {
		return nil, err
	}
	return &social.SetGroupAnnouncementResp{}, nil
}
