package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetMsgLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotGetMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetMsgLogic {
	return &BotGetMsgLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotGetMsgLogic) BotGetMsg(in *bot.BotGetMsgReq) (*bot.BotGetMsgResp, error) {
	return l.svcCtx.InstDao.GetMessage(l.ctx, in.MsgId, in.ConvId)
}