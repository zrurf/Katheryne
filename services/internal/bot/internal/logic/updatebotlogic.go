package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotLogic {
	return &UpdateBotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateBotLogic) UpdateBot(in *bot.UpdateBotReq) (*bot.UpdateBotResp, error) {
	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, in.BotId, in.Uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	updates := make(map[string]interface{})
	if in.Name != "" {
		updates["name"] = in.Name
	}
	if in.Avatar != "" {
		updates["avatar"] = in.Avatar
	}
	if in.Description != "" {
		updates["description"] = in.Description
	}
	if in.WebhookUrl != "" {
		updates["webhook_url"] = in.WebhookUrl
	}
	if len(in.SubscribeEvents) > 0 {
		updates["subscribe_events"] = in.SubscribeEvents
	}

	if len(updates) == 0 {
		return &bot.UpdateBotResp{}, nil
	}

	if err := l.svcCtx.BotDao.UpdateBot(l.ctx, in.BotId, in.Uid, updates); err != nil {
		return nil, err
	}

	return &bot.UpdateBotResp{}, nil
}