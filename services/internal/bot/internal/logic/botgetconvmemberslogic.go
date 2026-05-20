package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotGetConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvMembersLogic {
	return &BotGetConvMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotGetConvMembersLogic) BotGetConvMembers(in *bot.BotGetConvMembersReq) (*bot.BotGetConvMembersResp, error) {
	members, err := l.svcCtx.InstDao.GetGroupMembers(l.ctx, in.ConvId)
	if err != nil {
		return nil, err
	}

	return &bot.BotGetConvMembersResp{Members: members}, nil
}