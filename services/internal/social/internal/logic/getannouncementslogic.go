package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnnouncementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAnnouncementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnnouncementsLogic {
	return &GetAnnouncementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAnnouncementsLogic) GetAnnouncements(in *social.GetAnnouncementsReq) (*social.GetAnnouncementsResp, error) {
	// todo: add your logic here and delete this line

	return &social.GetAnnouncementsResp{}, nil
}
