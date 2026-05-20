package logic

import (
	"context"
	"fmt"
	"time"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotReplyMsgLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotReplyMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotReplyMsgLogic {
	return &BotReplyMsgLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotReplyMsgLogic) BotReplyMsg(in *bot.BotReplyMsgReq) (*bot.BotReplyMsgResp, error) {
	if !l.svcCtx.InstDao.IsInstalled(l.ctx, in.BotId, in.ConvId) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	msgID := time.Now().UnixNano()
	return &bot.BotReplyMsgResp{
		MsgId:     msgID,
		CreatedAt: time.Now().Unix(),
	}, nil
}