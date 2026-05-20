package developer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, req.BotID, uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	secretBytes := make([]byte, 16)
	rand.Read(secretBytes)
	newWebhookSecret := hex.EncodeToString(secretBytes)

	err = l.svcCtx.BotDao.UpdateWebhookSecret(l.ctx, req.BotID, newWebhookSecret)
	if err != nil {
		return nil, err
	}

	return &types.RotateWebhookSecretResp{WebhookSecret: newWebhookSecret}, nil
}
