package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyInstancesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyInstancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyInstancesLogic {
	return &ListMyInstancesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListMyInstancesLogic) ListMyInstances(in *bot.ListMyInstancesReq) (*bot.ListMyInstancesResp, error) {
	list, err := l.svcCtx.InstanceDao.ListInstancesByOwner(l.ctx, in.OwnerUid)
	if err != nil {
		return nil, err
	}
	return &bot.ListMyInstancesResp{List: list}, nil
}