package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotLogic {
	return &CreateBotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateBotLogic) CreateBot(in *bot.CreateBotReq) (*bot.CreateBotResp, error) {
	return l.svcCtx.BotDao.CreateBot(l.ctx, in.OwnerUid, in)
}