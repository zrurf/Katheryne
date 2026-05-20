package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotInstallationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBotInstallationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotInstallationsLogic {
	return &GetBotInstallationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBotInstallationsLogic) GetBotInstallations(in *bot.GetBotInstallationsReq) (*bot.GetBotInstallationsResp, error) {
	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, in.BotId, in.Uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	list, err := l.svcCtx.InstDao.ListBotInstallations(l.ctx, in.BotId)
	if err != nil {
		return nil, err
	}

	return &bot.GetBotInstallationsResp{List: list}, nil
}