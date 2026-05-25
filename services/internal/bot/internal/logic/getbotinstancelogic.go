package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotInstanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotInstanceLogic {
	return &GetBotInstanceLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetBotInstanceLogic) GetBotInstance(in *bot.GetBotInstanceReq) (*bot.GetBotInstanceResp, error) {
	var inst *bot.BotInstanceInfo
	var err error

	if in.InstanceId > 0 {
		inst, err = l.svcCtx.InstanceDao.GetInstanceByID(l.ctx, in.InstanceId)
	} else if in.BotId > 0 {
		inst, err = l.svcCtx.InstanceDao.GetInstanceByBotID(l.ctx, in.BotId)
	} else {
		return nil, fmt.Errorf("instance_id or bot_id required")
	}

	if err != nil {
		return nil, err
	}
	return &bot.GetBotInstanceResp{Instance: inst}, nil
}
