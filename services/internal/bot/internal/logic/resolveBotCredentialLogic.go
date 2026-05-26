package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveBotCredentialLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveBotCredentialLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveBotCredentialLogic {
	return &ResolveBotCredentialLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ResolveBotCredentialLogic) ResolveBotCredential(in *bot.ResolveBotCredentialReq) (*bot.ResolveBotCredentialResp, error) {
	botID, err := l.svcCtx.BotDao.ResolveByClientID(l.ctx, in.ClientId)
	if err != nil {
		return nil, fmt.Errorf("resolve bot credential: %w", err)
	}
	return &bot.ResolveBotCredentialResp{BotId: botID}, nil
}