package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBotTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBotTemplateLogic {
	return &DeleteBotTemplateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteBotTemplateLogic) DeleteBotTemplate(in *bot.DeleteBotTemplateReq) (*bot.DeleteBotTemplateResp, error) {
	if err := l.svcCtx.TemplateDao.DeleteTemplate(l.ctx, in.Uid, in.TemplateId); err != nil {
		return nil, err
	}
	return &bot.DeleteBotTemplateResp{}, nil
}