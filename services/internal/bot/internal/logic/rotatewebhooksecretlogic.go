package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RotateWebhookSecretLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRotateWebhookSecretLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RotateWebhookSecretLogic {
	return &RotateWebhookSecretLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RotateWebhookSecretLogic) RotateWebhookSecret(in *bot.RotateWebhookSecretReq) (*bot.RotateWebhookSecretResp, error) {
	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, in.BotId, in.Uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	secretBytes := make([]byte, 16)
	rand.Read(secretBytes)
	newWebhookSecret := hex.EncodeToString(secretBytes)

	err := l.svcCtx.BotDao.UpdateWebhookSecret(l.ctx, in.BotId, newWebhookSecret)
	if err != nil {
		return nil, err
	}

	return &bot.RotateWebhookSecretResp{WebhookSecret: newWebhookSecret}, nil
}