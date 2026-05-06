// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package installation

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InstallBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InstallBotLogic {
	return &InstallBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InstallBotLogic) InstallBot(req *types.InstallBotReq) (resp *types.InstallBotResp, err error) {
	// todo: add your logic here and delete this line

	return
}
