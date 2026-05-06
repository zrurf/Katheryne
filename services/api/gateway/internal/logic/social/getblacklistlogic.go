// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlacklistLogic {
	return &GetBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlacklistLogic) GetBlacklist() (resp *types.GetBlacklistResp, err error) {
	// todo: add your logic here and delete this line

	return
}
