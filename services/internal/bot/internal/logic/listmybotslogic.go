package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyBotsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyBotsLogic {
	return &ListMyBotsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListMyBotsLogic) ListMyBots(in *bot.ListMyBotsReq) (*bot.ListMyBotsResp, error) {
	list, err := l.svcCtx.BotDao.ListBotsByOwner(l.ctx, in.OwnerUid)
	if err != nil {
		return nil, err
	}

	return &bot.ListMyBotsResp{List: list}, nil
}