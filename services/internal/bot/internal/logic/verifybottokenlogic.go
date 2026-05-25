package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyBotTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyBotTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyBotTokenLogic {
	return &VerifyBotTokenLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *VerifyBotTokenLogic) VerifyBotToken(in *bot.VerifyBotTokenReq) (*bot.VerifyBotTokenResp, error) {
	botID, instanceID, isOfficial, err := l.svcCtx.InstanceDao.VerifyBotToken(l.ctx, in.BotToken)
	if err != nil {
		return &bot.VerifyBotTokenResp{Valid: false}, nil
	}
	return &bot.VerifyBotTokenResp{
		Valid:      true,
		BotId:      botID,
		InstanceId: instanceID,
		IsOfficial: isOfficial,
	}, nil
}