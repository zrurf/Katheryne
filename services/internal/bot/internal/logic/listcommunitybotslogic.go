package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommunityBotsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCommunityBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommunityBotsLogic {
	return &ListCommunityBotsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListCommunityBotsLogic) ListCommunityBots(in *bot.ListCommunityBotsReq) (*bot.ListCommunityBotsResp, error) {
	list, err := l.svcCtx.BotDao.ListCommunityBots(l.ctx, in.Keyword)
	if err != nil {
		return nil, err
	}

	return &bot.ListCommunityBotsResp{List: list}, nil
}