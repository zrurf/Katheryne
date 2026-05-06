// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetUserLogic {
	return &BotGetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetUserLogic) BotGetUser(req *types.BotGetUserReq) (resp *types.BotGetUserResp, err error) {
	// todo: add your logic here and delete this line

	return
}
