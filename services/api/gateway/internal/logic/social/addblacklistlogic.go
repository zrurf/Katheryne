// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBlacklistLogic {
	return &AddBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddBlacklistLogic) AddBlacklist(req *types.AddBlacklistReq) (resp *types.AddBlacklistResp, err error) {
	// todo: add your logic here and delete this line

	return
}
