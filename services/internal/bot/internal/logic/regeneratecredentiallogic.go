package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegenerateCredentialLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegenerateCredentialLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegenerateCredentialLogic {
	return &RegenerateCredentialLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegenerateCredentialLogic) RegenerateCredential(in *bot.RegenerateCredentialReq) (*bot.RegenerateCredentialResp, error) {
	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, in.BotId, in.Uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	clientID, clientSecret, err := l.svcCtx.BotDao.RegenerateCredential(l.ctx, in.BotId)
	if err != nil {
		return nil, err
	}

	return &bot.RegenerateCredentialResp{
		ClientId:     clientID,
		ClientSecret: clientSecret,
	}, nil
}