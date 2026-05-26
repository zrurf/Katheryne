package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotInstanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotInstanceLogic {
	return &CreateBotInstanceLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateBotInstanceLogic) CreateBotInstance(in *bot.CreateBotInstanceReq) (*bot.CreateBotInstanceResp, error) {
	// Let DAO handle the hosted_by logic properly.
	// Do NOT force IsSelfHosted=true — orchestrator relies on is_self_hosted=FALSE
	// to discover platform-hosted instances.
	return l.svcCtx.InstanceDao.CreateInstance(l.ctx, in)
}
