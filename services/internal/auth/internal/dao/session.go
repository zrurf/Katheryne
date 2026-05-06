package dao

import (
	"auth/internal/model"
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	AccessPrefix       = "access_token:"
	RefreshPrefix      = "refresh_token:"
	TFAPrefix          = "2fa_token:"
	LoginSessionPrefix = "login_session:"
	TokenMapPrefix     = "token_map:"
)

type SessionDao struct {
	redis *redis.Client
}

func NewSessionDao(redis *redis.Client) *SessionDao {
	return &SessionDao{
		redis: redis,
	}
}

// SaveAccessToken 保存访问令牌
func (r *SessionDao) SaveAccessToken(ctx context.Context, token string, uid int64, expireSec int) error {
	key := AccessPrefix + token
	return r.redis.Set(ctx, key, uid, time.Second*time.Duration(expireSec)).Err()
}

// SaveRefreshToken 保存刷新令牌
func (r *SessionDao) SaveRefreshToken(ctx context.Context, token string, uid int64, expireSec int) error {
	key := RefreshPrefix + token
	return r.redis.Set(ctx, key, uid, time.Second*time.Duration(expireSec)).Err()
}

// SaveTokenMap 保存Token映射
func (r *SessionDao) SaveTokenMap(ctx context.Context, accessToken string, refreshToken string, expireSec int) error {
	key := TokenMapPrefix + refreshToken
	return r.redis.Set(ctx, key, accessToken, time.Second*time.Duration(expireSec)).Err()
}

// Save2FAToken 保存2FA令牌
func (r *SessionDao) Save2FAToken(ctx context.Context, token string, uid int64, expireSec int) error {
	key := TFAPrefix + token
	return r.redis.Set(ctx, key, uid, time.Second*time.Duration(expireSec)).Err()
}

// SaveLoginSession 保存登录Session
func (r *SessionDao) SaveLoginSession(ctx context.Context, sessionId string, sessionData *model.LoginSession, expireSec int) error {
	key := "login_session:" + sessionId
	data, err := sonic.Marshal(sessionData)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, key, data, time.Second*time.Duration(expireSec)).Err()
}

// GetUidByAccessToken 通过访问令牌获取用户ID
func (r *SessionDao) GetUidByAccessToken(ctx context.Context, token string) (int64, error) {
	return r.redis.Get(ctx, AccessPrefix+token).Int64()
}

// GetUidByRefreshToken 通过刷新令牌获取用户ID
func (r *SessionDao) GetUidByRefreshToken(ctx context.Context, token string) (int64, error) {
	return r.redis.Get(ctx, RefreshPrefix+token).Int64()
}

// GetUidBy2FAToken 通过2FA令牌获取用户ID
func (r *SessionDao) GetUidBy2FAToken(ctx context.Context, token string) (int64, error) {
	return r.redis.Get(ctx, TFAPrefix+token).Int64()
}

// GetLoginSession 获取登录会话
func (r *SessionDao) GetLoginSession(ctx context.Context, sessionId string) (*model.LoginSession, error) {
	v, err := r.redis.Get(ctx, LoginSessionPrefix+sessionId).Result()
	if err != nil {
		return nil, err
	}
	var result model.LoginSession
	if err := sonic.Unmarshal([]byte(v), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTokenMap 获取TokenMap
func (r *SessionDao) GetTokenMap(ctx context.Context, refreshToken string) (string, error) {
	key := TokenMapPrefix + refreshToken
	return r.redis.Get(ctx, key).Result()
}

// DelAccessToken 删除访问令牌
func (r *SessionDao) DelAccessToken(ctx context.Context, token string) error {
	return r.redis.Del(ctx, AccessPrefix+token).Err()
}

// DelRefreshToken 删除刷新令牌
func (r *SessionDao) DelRefreshToken(ctx context.Context, token string) error {
	return r.redis.Del(ctx, RefreshPrefix+token).Err()
}

// Del2FAToken 删除2FA令牌
func (r *SessionDao) Del2FAToken(ctx context.Context, token string) error {
	return r.redis.Del(ctx, TFAPrefix+token).Err()
}

// DelTokenMap 删除令牌映射
func (r *SessionDao) DelTokenMap(ctx context.Context, refreshToken string) error {
	return r.redis.Del(ctx, TokenMapPrefix+refreshToken).Err()
}

// DelLoginSession 删除登录会话
func (r *SessionDao) DelLoginSession(ctx context.Context, sessionId string) error {
	return r.redis.Del(ctx, LoginSessionPrefix+sessionId).Err()
}

// HasAccessToken 检查访问令牌是否存在
func (r *SessionDao) HasAccessToken(ctx context.Context, token string) (bool, error) {
	res, err := r.redis.Exists(ctx, AccessPrefix+token).Result()
	return res != 0, err
}

// HasRefreshToken 检查刷新令牌是否存在
func (r *SessionDao) HasRefreshToken(ctx context.Context, token string) (bool, error) {
	res, err := r.redis.Exists(ctx, RefreshPrefix+token).Result()
	return res != 0, err
}

// Has2FAToken 检查2FA令牌是否存在
func (r *SessionDao) Has2FAToken(ctx context.Context, token string) (bool, error) {
	res, err := r.redis.Exists(ctx, TFAPrefix+token).Result()
	return res != 0, err
}

// HasLoginSession 检查登录会话是否存在
func (r *SessionDao) HasLoginSession(ctx context.Context, token string) (bool, error) {
	res, err := r.redis.Exists(ctx, LoginSessionPrefix+token).Result()
	return res != 0, err
}

// HasTokenMap 检查TokenMap是否存在
func (r *SessionDao) HasTokenMap(ctx context.Context, refreshToken string) (bool, error) {
	res, err := r.redis.Exists(ctx, TokenMapPrefix+refreshToken).Result()
	return res != 0, err
}
