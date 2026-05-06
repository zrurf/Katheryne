// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package conversation

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearUnreadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearUnreadLogic {
	return &ClearUnreadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearUnreadLogic) ClearUnread(req *types.ClearUnreadReq) (resp *types.ClearUnreadResp, err error) {
	// todo: add your logic here and delete this line

	return
}
