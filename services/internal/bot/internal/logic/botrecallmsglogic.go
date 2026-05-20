package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotRecallMsgLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotRecallMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotRecallMsgLogic {
	return &BotRecallMsgLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotRecallMsgLogic) BotRecallMsg(in *bot.BotRecallMsgReq) (*bot.BotRecallMsgResp, error) {
	if !l.svcCtx.InstDao.IsInstalled(l.ctx, in.BotId, in.ConvId) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	return &bot.BotRecallMsgResp{}, nil
}