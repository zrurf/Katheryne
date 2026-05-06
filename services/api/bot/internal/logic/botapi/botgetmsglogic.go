// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetMsgLogic {
	return &BotGetMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetMsgLogic) BotGetMsg(req *types.BotGetMsgReq) (resp *types.BotGetMsgResp, err error) {
	// todo: add your logic here and delete this line

	return
}
