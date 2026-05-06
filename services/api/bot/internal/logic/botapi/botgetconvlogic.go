// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvLogic {
	return &BotGetConvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetConvLogic) BotGetConv(req *types.BotGetConvReq) (resp *types.BotGetConvResp, err error) {
	// todo: add your logic here and delete this line

	return
}
