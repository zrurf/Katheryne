// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotRecallMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotRecallMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotRecallMsgLogic {
	return &BotRecallMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotRecallMsgLogic) BotRecallMsg(req *types.BotRecallMsgReq) (resp *types.BotRecallMsgResp, err error) {
	// todo: add your logic here and delete this line

	return
}
