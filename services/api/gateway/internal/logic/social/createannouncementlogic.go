// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAnnouncementLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAnnouncementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAnnouncementLogic {
	return &CreateAnnouncementLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAnnouncementLogic) CreateAnnouncement(req *types.CreateAnnouncementReq) (resp *types.CreateAnnouncementResp, err error) {
	// todo: add your logic here and delete this line

	return
}
