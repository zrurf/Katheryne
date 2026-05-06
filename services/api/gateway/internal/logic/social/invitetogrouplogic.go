// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InviteToGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInviteToGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteToGroupLogic {
	return &InviteToGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InviteToGroupLogic) InviteToGroup(req *types.InviteToGroupReq) (resp *types.InviteToGroupResp, err error) {
	// todo: add your logic here and delete this line

	return
}
