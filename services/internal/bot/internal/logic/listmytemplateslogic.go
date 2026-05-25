package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyTemplatesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyTemplatesLogic {
	return &ListMyTemplatesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListMyTemplatesLogic) ListMyTemplates(in *bot.ListMyTemplatesReq) (*bot.ListMyTemplatesResp, error) {
	list, err := l.svcCtx.TemplateDao.ListTemplatesByOwner(l.ctx, in.OwnerUid)
	if err != nil {
		return nil, err
	}
	return &bot.ListMyTemplatesResp{List: list}, nil
}