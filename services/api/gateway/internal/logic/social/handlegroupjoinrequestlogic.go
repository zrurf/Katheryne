// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupJoinRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleGroupJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupJoinRequestLogic {
	return &HandleGroupJoinRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGroupJoinRequestLogic) HandleGroupJoinRequest(req *types.HandleGroupJoinReq) (resp *types.HandleGroupJoinResp, err error) {
	// todo: add your logic here and delete this line

	return
}
