package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBotInstanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBotInstanceLogic {
	return &DeleteBotInstanceLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteBotInstanceLogic) DeleteBotInstance(in *bot.DeleteBotInstanceReq) (*bot.DeleteBotInstanceResp, error) {
	if err := l.svcCtx.InstanceDao.DeleteInstance(l.ctx, in.Uid, in.InstanceId); err != nil {
		return nil, err
	}
	return &bot.DeleteBotInstanceResp{}, nil
}