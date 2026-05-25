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
	// Set self-hosted by default if not specified
	if !in.IsSelfHosted && in.HostedBy == 0 {
		in.IsSelfHosted = true
	}
	return l.svcCtx.InstanceDao.CreateInstance(l.ctx, in)
}