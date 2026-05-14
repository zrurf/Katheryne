package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupAnnouncementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupAnnouncementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupAnnouncementLogic {
	return &GetGroupAnnouncementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupAnnouncementLogic) GetGroupAnnouncement(in *social.GetGroupAnnouncementReq) (*social.GetGroupAnnouncementResp, error) {
	announcements, _, err := l.svcCtx.SocialDao.ListAnnouncements(l.ctx, in.GroupId, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(announcements) == 0 {
		return &social.GetGroupAnnouncementResp{Content: ""}, nil
	}
	return &social.GetGroupAnnouncementResp{Content: announcements[0].Content}, nil
}
