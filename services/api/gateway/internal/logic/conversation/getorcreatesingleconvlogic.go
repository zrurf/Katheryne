// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package conversation

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrCreateSingleConvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrCreateSingleConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateSingleConvLogic {
	return &GetOrCreateSingleConvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrCreateSingleConvLogic) GetOrCreateSingleConv(req *types.GetOrCreateSingleConvReq) (resp *types.GetOrCreateSingleConvResp, err error) {
	// todo: add your logic here and delete this line

	return
}
