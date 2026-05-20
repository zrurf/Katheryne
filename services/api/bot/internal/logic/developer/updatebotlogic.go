package developer

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotLogic {
	return &UpdateBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBotLogic) UpdateBot(req *types.UpdateBotReq) (resp *types.UpdateBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, req.BotID, uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.WebhookURL != "" {
		updates["webhook_url"] = req.WebhookURL
	}
	if len(req.SubscribeEvents) > 0 {
		updates["subscribe_events"] = req.SubscribeEvents
	}

	if len(updates) == 0 {
		return &types.UpdateBotResp{}, nil
	}

	if err := l.svcCtx.BotDao.UpdateBot(l.ctx, req.BotID, uid, updates); err != nil {
		return nil, err
	}

	return &types.UpdateBotResp{}, nil
}