// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotReplyMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotReplyMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotReplyMsgLogic {
	return &BotReplyMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotReplyMsgLogic) BotReplyMsg(req *types.BotReplyMsgReq) (resp *types.BotReplyMsgResp, err error) {
	// todo: add your logic here and delete this line

	return
}
