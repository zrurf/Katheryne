package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConvBotsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConvBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvBotsLogic {
	return &GetConvBotsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConvBotsLogic) GetConvBots(in *bot.GetConvBotsReq) (*bot.GetConvBotsResp, error) {
	list, err := l.svcCtx.InstDao.ListConvBots(l.ctx, in.ConvId)
	if err != nil {
		return &bot.GetConvBotsResp{List: []*bot.InstalledBotItem{}}, nil
	}

	return &bot.GetConvBotsResp{List: list}, nil
}