package dao

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type OAuthDao struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewOAuthDao(db *pgxpool.Pool, redis *redis.Client) *OAuthDao {
	return &OAuthDao{db: db, redis: redis}
}

func (d *OAuthDao) StoreAuthCode(ctx context.Context, clientID, redirectURI, scope string, uid, convID int64) (string, error) {
	codeBytes := make([]byte, 32)
	rand.Read(codeBytes)
	code := hex.EncodeToString(codeBytes)

	_, err := d.db.Exec(ctx,
		`INSERT INTO bot_auth_code (code, client_id, redirect_uri, scope, uid, conv_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		code, clientID, redirectURI, scope, uid, convID,
		time.Now().Add(10*time.Minute))
	if err != nil {
		return "", err
	}

	authData := map[string]interface{}{
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"scope":        scope,
		"conv_id":      convID,
		"uid":          uid,
	}
	data, _ := json.Marshal(authData)
	d.redis.Set(ctx, "oauth2:code:"+code, data, 10*time.Minute)

	return code, nil
}

func (d *OAuthDao) GetAuthCode(ctx context.Context, code string) (map[string]interface{}, error) {
	data, err := d.redis.Get(ctx, "oauth2:code:"+code).Result()
	if err != nil {
		return nil, err
	}

	var authData map[string]interface{}
	json.Unmarshal([]byte(data), &authData)
	return authData, nil
}

func (d *OAuthDao) ConsumeAuthCode(ctx context.Context, code string) error {
	return d.redis.Del(ctx, "oauth2:code:"+code).Err()
}

func (d *OAuthDao) StoreAccessToken(ctx context.Context, accessToken string, botID int64, clientID, scope string, expiresIn int64) error {
	tokenInfo := map[string]interface{}{
		"bot_id":     botID,
		"client_id":  clientID,
		"scope":      scope,
		"expires_at": time.Now().Unix() + expiresIn,
	}
	tokenData, _ := json.Marshal(tokenInfo)
	return d.redis.Set(ctx, "bot_access_token:"+accessToken, tokenData, time.Duration(expiresIn)*time.Second).Err()
}

func (d *OAuthDao) StoreRefreshToken(ctx context.Context, refreshToken, clientID, scope string) error {
	d.redis.Set(ctx, "oauth2:refresh_token:"+refreshToken, clientID, 30*24*time.Hour)
	d.redis.Set(ctx, "oauth2:token_scope:"+refreshToken, scope, 30*24*time.Hour)
	return nil
}

func (d *OAuthDao) ConsumeRefreshToken(ctx context.Context, refreshToken string) (clientID, scope string, err error) {
	clientID, err = d.redis.Get(ctx, "oauth2:refresh_token:"+refreshToken).Result()
	if err != nil {
		return "", "", err
	}

	scope, _ = d.redis.Get(ctx, "oauth2:token_scope:"+refreshToken).Result()

	d.redis.Del(ctx, "oauth2:refresh_token:"+refreshToken)
	d.redis.Del(ctx, "oauth2:token_scope:"+refreshToken)

	return clientID, scope, nil
}

func (d *OAuthDao) AddConvBotCache(ctx context.Context, convID, botID int64) {
	key := "conv_bots:" + strconv.FormatInt(convID, 10)
	d.redis.SAdd(ctx, key, botID)
}

func (d *OAuthDao) RemoveConvBotCache(ctx context.Context, convID, botID int64) {
	key := "conv_bots:" + strconv.FormatInt(convID, 10)
	d.redis.SRem(ctx, key, botID)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}