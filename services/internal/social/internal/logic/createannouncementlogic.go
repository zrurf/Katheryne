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

// 群公告
func (l *CreateAnnouncementLogic) CreateAnnouncement(in *social.CreateAnnouncementReq) (*social.CreateAnnouncementResp, error) {
	// todo: add your logic here and delete this line

	return &social.CreateAnnouncementResp{}, nil
}
