package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotGetConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvLogic {
	return &BotGetConvLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotGetConvLogic) BotGetConv(in *bot.BotGetConvReq) (*bot.BotGetConvResp, error) {
	convType, name, avatar, groupID, createdAt, err := l.svcCtx.InstDao.GetConvInfo(l.ctx, in.ConvId)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	resp := &bot.BotGetConvResp{
		ConvId:    in.ConvId,
		Type:      convType,
		Name:      name,
		Avatar:    avatar,
		CreatedAt: createdAt,
	}
	if groupID > 0 {
		resp.GroupId = groupID
	}

	return resp, nil
}