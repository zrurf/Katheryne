package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotLogic {
	return &GetBotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBotLogic) GetBot(in *bot.GetBotReq) (*bot.GetBotResp, error) {
	botInfo, err := l.svcCtx.BotDao.GetBotByID(l.ctx, in.BotId, in.Uid)
	if err != nil {
		return nil, err
	}

	return &bot.GetBotResp{Bot: botInfo}, nil
}