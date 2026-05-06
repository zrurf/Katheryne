// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransferOwnerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTransferOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferOwnerLogic {
	return &TransferOwnerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TransferOwnerLogic) TransferOwner(req *types.TransferOwnerReq) (resp *types.TransferOwnerResp, err error) {
	// todo: add your logic here and delete this line

	return
}
