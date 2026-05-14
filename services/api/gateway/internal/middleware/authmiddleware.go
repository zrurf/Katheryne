// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
)

type AuthMiddleware struct {
	redis *redis.Client
}

func NewAuthMiddleware(redisClient *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		redis: redisClient,
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token := parts[1]
		uid, err := m.redis.Get(r.Context(), "access_token:"+token).Int64()
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "uid", uid)
		next(w, r.WithContext(ctx))
	}
}
