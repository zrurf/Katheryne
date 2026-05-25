package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishBotTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishBotTemplateLogic {
	return &PublishBotTemplateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *PublishBotTemplateLogic) PublishBotTemplate(in *bot.PublishBotTemplateReq) (*bot.PublishBotTemplateResp, error) {
	if err := l.svcCtx.TemplateDao.PublishTemplate(l.ctx, in.Uid, in.TemplateId); err != nil {
		return nil, err
	}
	return &bot.PublishBotTemplateResp{}, nil
}