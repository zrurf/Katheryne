// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotInstallationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotInstallationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotInstallationsLogic {
	return &GetBotInstallationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotInstallationsLogic) GetBotInstallations(req *types.GetBotInstallationsReq) (resp *types.GetBotInstallationsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
