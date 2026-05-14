package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOnlineStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOnlineStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOnlineStatusLogic {
	return &UpdateOnlineStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateOnlineStatusLogic) UpdateOnlineStatus(in *social.UpdateOnlineStatusReq) (*social.UpdateOnlineStatusResp, error) {
	return &social.UpdateOnlineStatusResp{}, nil
}