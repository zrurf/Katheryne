package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotTemplateLogic {
	return &CreateBotTemplateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateBotTemplateLogic) CreateBotTemplate(in *bot.CreateBotTemplateReq) (*bot.CreateBotTemplateResp, error) {
	templateID, err := l.svcCtx.TemplateDao.CreateTemplate(l.ctx, in)
	if err != nil {
		return nil, err
	}
	return &bot.CreateBotTemplateResp{TemplateId: templateID}, nil
}
