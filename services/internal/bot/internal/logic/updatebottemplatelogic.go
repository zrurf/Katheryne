package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotTemplateLogic {
	return &UpdateBotTemplateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateBotTemplateLogic) UpdateBotTemplate(in *bot.UpdateBotTemplateReq) (*bot.UpdateBotTemplateResp, error) {
	if err := l.svcCtx.TemplateDao.UpdateTemplate(l.ctx, in.Uid, in.TemplateId, in); err != nil {
		return nil, err
	}
	return &bot.UpdateBotTemplateResp{}, nil
}