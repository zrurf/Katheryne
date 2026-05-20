package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type BotAuthMiddleware struct {
	redis *redis.Client
}

func NewBotAuthMiddleware(redisClient *redis.Client) *BotAuthMiddleware {
	return &BotAuthMiddleware{
		redis: redisClient,
	}
}

type BotTokenInfo struct {
	BotID      int64  `json:"bot_id"`
	ClientID   string `json:"client_id"`
	Scope      string `json:"scope"`
	ExpiresAt  int64  `json:"expires_at"`
	InstallUID int64  `json:"install_uid,omitempty"`
}

func (m *BotAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		tokenData, err := m.redis.Get(r.Context(), "bot_access_token:"+token).Result()
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var botInfo BotTokenInfo
		if err := json.Unmarshal([]byte(tokenData), &botInfo); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if botInfo.ExpiresAt < time.Now().Unix() {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "bot_id", botInfo.BotID)
		ctx = context.WithValue(ctx, "client_id", botInfo.ClientID)
		ctx = context.WithValue(ctx, "scope", botInfo.Scope)

		authCtx := &BotAuthContext{
			BotID:     botInfo.BotID,
			ClientID:  botInfo.ClientID,
			Scope:     botInfo.Scope,
			Token:     token,
		}
		ctx = context.WithValue(ctx, "bot_auth", authCtx)

		next(w, r.WithContext(ctx))
	}
}

type BotAuthContext struct {
	BotID     int64
	ClientID  string
	Scope     string
	Token     string
}

func GetBotAuth(ctx context.Context) (*BotAuthContext, error) {
	auth, ok := ctx.Value("bot_auth").(*BotAuthContext)
	if !ok {
		return nil, fmt.Errorf("bot auth context not found")
	}
	return auth, nil
}

func GetBotID(ctx context.Context) int64 {
	botID, _ := ctx.Value("bot_id").(int64)
	return botID
}
