// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package installation

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UninstallBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUninstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UninstallBotLogic {
	return &UninstallBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UninstallBotLogic) UninstallBot(req *types.UninstallBotReq) (resp *types.UninstallBotResp, err error) {
	// todo: add your logic here and delete this line

	return
}
