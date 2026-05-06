// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyBotsLogic {
	return &ListMyBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyBotsLogic) ListMyBots(req *types.ListMyBotsResp) (resp *types.ListMyBotsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
