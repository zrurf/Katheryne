// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupInviteLogic {
	return &HandleGroupInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGroupInviteLogic) HandleGroupInvite(req *types.HandleGroupInviteReq) (resp *types.HandleGroupInviteResp, err error) {
	// todo: add your logic here and delete this line

	return
}
