package logic

import (
	"context"
	"fmt"
	"time"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotSendMsgLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotSendMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotSendMsgLogic {
	return &BotSendMsgLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotSendMsgLogic) BotSendMsg(in *bot.BotSendMsgReq) (*bot.BotSendMsgResp, error) {
	if !l.svcCtx.InstDao.IsInstalled(l.ctx, in.BotId, in.ConvId) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	msgID := time.Now().UnixNano()
	return &bot.BotSendMsgResp{
		MsgId:     msgID,
		CreatedAt: time.Now().Unix(),
	}, nil
}