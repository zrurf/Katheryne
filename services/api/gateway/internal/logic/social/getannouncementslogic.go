// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnnouncementsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAnnouncementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnnouncementsLogic {
	return &GetAnnouncementsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAnnouncementsLogic) GetAnnouncements(req *types.GetAnnouncementsReq) (resp *types.GetAnnouncementsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
