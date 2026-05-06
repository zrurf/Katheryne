package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInvitesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupInvitesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInvitesLogic {
	return &GetGroupInvitesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupInvitesLogic) GetGroupInvites(in *social.GetGroupInvitesReq) (*social.GetGroupInvitesResp, error) {
	// todo: add your logic here and delete this line

	return &social.GetGroupInvitesResp{}, nil
}
