package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotInstanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotInstanceLogic {
	return &UpdateBotInstanceLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateBotInstanceLogic) UpdateBotInstance(in *bot.UpdateBotInstanceReq) (*bot.UpdateBotInstanceResp, error) {
	if err := l.svcCtx.InstanceDao.UpdateInstance(l.ctx, in.Uid, in.InstanceId, in); err != nil {
		return nil, err
	}
	return &bot.UpdateBotInstanceResp{}, nil
}