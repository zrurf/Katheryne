package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListHostedInstancesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListHostedInstancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListHostedInstancesLogic {
	return &ListHostedInstancesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListHostedInstancesLogic) ListHostedInstances(in *bot.ListHostedInstancesReq) (*bot.ListHostedInstancesResp, error) {
	instances, err := l.svcCtx.InstanceDao.ListHostedInstances(l.ctx, "")
	if err != nil {
		return nil, err
	}
	return &bot.ListHostedInstancesResp{List: instances}, nil
}