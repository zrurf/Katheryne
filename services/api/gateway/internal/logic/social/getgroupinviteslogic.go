// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInvitesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupInvitesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInvitesLogic {
	return &GetGroupInvitesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupInvitesLogic) GetGroupInvites(req *types.GetGroupInvitesReq) (resp *types.GetGroupInvitesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
