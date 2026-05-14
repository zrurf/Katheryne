package developer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotLogic {
	return &CreateBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateBotLogic) CreateBot(req *types.CreateBotReq) (resp *types.CreateBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	botID := time.Now().UnixNano()
	clientID := "bot_" + randomHex(16)
	clientSecret := randomHex(32)

	bot := map[string]interface{}{
		"bot_id":           botID,
		"name":             req.Name,
		"avatar":           req.Avatar,
		"description":      req.Description,
		"owner_uid":        uid,
		"webhook_url":      req.WebhookURL,
		"subscribe_events": req.SubscribeEvents,
		"status":           "active",
		"client_id":        clientID,
		"client_secret":    clientSecret,
		"created_at":       time.Now().Unix(),
	}

	data, _ := json.Marshal(bot)
	l.svcCtx.Redis.HSet(l.ctx, "bots", fmt.Sprintf("%d", botID), data)
	l.svcCtx.Redis.SAdd(l.ctx, fmt.Sprintf("user_bots:%d", uid), botID)

	return &types.CreateBotResp{
		BotID:        botID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}