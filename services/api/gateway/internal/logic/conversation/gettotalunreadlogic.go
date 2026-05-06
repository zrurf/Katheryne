// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package conversation

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTotalUnreadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTotalUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTotalUnreadLogic {
	return &GetTotalUnreadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTotalUnreadLogic) GetTotalUnread() (resp *types.GetTotalUnreadResp, err error) {
	// todo: add your logic here and delete this line

	return
}
