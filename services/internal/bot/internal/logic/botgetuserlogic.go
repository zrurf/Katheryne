package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetUserLogic {
	return &BotGetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotGetUserLogic) BotGetUser(in *bot.BotGetUserReq) (*bot.BotGetUserResp, error) {
	name, avatar, err := l.svcCtx.InstDao.GetUserInfo(l.ctx, in.Uid)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &bot.BotGetUserResp{
		Uid:    in.Uid,
		Name:   name,
		Avatar: avatar,
	}, nil
}