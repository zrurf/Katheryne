package developer

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RotateWebhookSecretLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRotateWebhookSecretLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RotateWebhookSecretLogic {
	return &RotateWebhookSecretLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RotateWebhookSecretLogic) RotateWebhookSecret(req *types.RotateWebhookSecretReq) (resp *types.RotateWebhookSecretResp, err error) {
	data, err := l.svcCtx.Redis.HGet(l.ctx, "bots", fmt.Sprintf("%d", req.BotID)).Result()
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	secret := randomHex(32)

	var bot map[string]interface{}
	json.Unmarshal([]byte(data), &bot)
	bot["webhook_secret"] = secret

	data2, _ := json.Marshal(bot)
	l.svcCtx.Redis.HSet(l.ctx, "bots", fmt.Sprintf("%d", req.BotID), data2)

	return &types.RotateWebhookSecretResp{
		WebhookSecret: secret,
	}, nil
}
