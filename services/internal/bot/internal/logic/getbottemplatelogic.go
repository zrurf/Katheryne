package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotTemplateLogic {
	return &GetBotTemplateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetBotTemplateLogic) GetBotTemplate(in *bot.GetBotTemplateReq) (*bot.GetBotTemplateResp, error) {
	tmpl, err := l.svcCtx.TemplateDao.GetTemplateByID(l.ctx, in.TemplateId)
	if err != nil {
		return nil, err
	}
	return &bot.GetBotTemplateResp{Template: tmpl}, nil
}