package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBotLogic {
	return &DeleteBotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteBotLogic) DeleteBot(in *bot.DeleteBotReq) (*bot.DeleteBotResp, error) {
	rowsAffected, err := l.svcCtx.BotDao.DeleteBot(l.ctx, in.BotId, in.Uid)
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	return &bot.DeleteBotResp{}, nil
}