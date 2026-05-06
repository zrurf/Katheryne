// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oauth2

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveAuthorizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveAuthorizeLogic {
	return &ApproveAuthorizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveAuthorizeLogic) ApproveAuthorize(req *types.ApproveAuthorizeReq) (resp *types.ApproveAuthorizeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
