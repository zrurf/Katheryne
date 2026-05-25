package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotRuntimeConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBotRuntimeConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotRuntimeConfigLogic {
	return &GetBotRuntimeConfigLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetBotRuntimeConfigLogic) GetBotRuntimeConfig(in *bot.GetBotRuntimeConfigReq) (*bot.GetBotRuntimeConfigResp, error) {
	config, err := l.svcCtx.InstanceDao.GetRuntimeConfig(l.ctx, in.BotId, in.ConvId)
	if err != nil {
		return nil, err
	}
	return config, nil
}