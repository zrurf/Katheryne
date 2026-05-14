package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAnnouncementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAnnouncementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAnnouncementLogic {
	return &CreateAnnouncementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAnnouncementLogic) CreateAnnouncement(in *social.CreateAnnouncementReq) (*social.CreateAnnouncementResp, error) {
	_, err := l.svcCtx.SocialDao.InsertAnnouncement(l.ctx, in.GroupId, in.Uid, in.Content)
	if err != nil {
		return nil, err
	}
	return &social.CreateAnnouncementResp{}, nil
}
