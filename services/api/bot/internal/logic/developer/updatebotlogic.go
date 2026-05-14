package developer

import (
	"context"
	"encoding/json"
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
	data, err := l.svcCtx.Redis.HGet(l.ctx, "bots", fmt.Sprintf("%d", req.BotID)).Result()
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	var bot map[string]interface{}
	json.Unmarshal([]byte(data), &bot)

	if req.Name != "" {
		bot["name"] = req.Name
	}
	if req.Avatar != "" {
		bot["avatar"] = req.Avatar
	}
	if req.Description != "" {
		bot["description"] = req.Description
	}
	if req.WebhookURL != "" {
		bot["webhook_url"] = req.WebhookURL
	}
	if req.SubscribeEvents != nil {
		bot["subscribe_events"] = req.SubscribeEvents
	}

	data2, _ := json.Marshal(bot)
	l.svcCtx.Redis.HSet(l.ctx, "bots", fmt.Sprintf("%d", req.BotID), data2)

	return &types.UpdateBotResp{}, nil
}