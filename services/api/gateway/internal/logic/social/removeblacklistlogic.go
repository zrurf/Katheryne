// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveBlacklistLogic {
	return &RemoveBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveBlacklistLogic) RemoveBlacklist(req *types.RemoveBlacklistReq) (resp *types.RemoveBlacklistResp, err error) {
	// todo: add your logic here and delete this line

	return
}
