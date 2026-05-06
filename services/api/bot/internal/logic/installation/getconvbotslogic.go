// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package installation

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConvBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConvBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvBotsLogic {
	return &GetConvBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConvBotsLogic) GetConvBots(req *types.GetConvBotsReq) (resp *types.GetConvBotsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
