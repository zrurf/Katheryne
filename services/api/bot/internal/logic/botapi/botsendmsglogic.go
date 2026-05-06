// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotSendMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotSendMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotSendMsgLogic {
	return &BotSendMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotSendMsgLogic) BotSendMsg(req *types.BotSendMsgReq) (resp *types.BotSendMsgResp, err error) {
	// todo: add your logic here and delete this line

	return
}
