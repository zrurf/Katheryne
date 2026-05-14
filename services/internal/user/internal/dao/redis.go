package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	UserCachePrefix       = "user:info:"
	UserSettingsPrefix    = "user:settings:"
	UserBanPrefix         = "user:ban:"
	UserDevicesPrefix     = "user:devices:"
	CacheTTL              = 10 * time.Minute
)

// RedisDao Redis 缓存访问层
type RedisDao struct {
	rdb *redis.Client
}

func NewRedisDao(rdb *redis.Client) *RedisDao {
	return &RedisDao{rdb: rdb}
}

func (r *RedisDao) keyUser(uid int64) string {
	return fmt.Sprintf("%s%d", UserCachePrefix, uid)
}

func (r *RedisDao) keySettings(uid int64) string {
	return fmt.Sprintf("%s%d", UserSettingsPrefix, uid)
}

func (r *RedisDao) keyBan(uid int64) string {
	return fmt.Sprintf("%s%d", UserBanPrefix, uid)
}

func (r *RedisDao) keyDevices(uid int64) string {
	return fmt.Sprintf("%s%d", UserDevicesPrefix, uid)
}

// ---------- 用户信息缓存 ----------

// SetUserCache 缓存用户信息
func (r *RedisDao) SetUserCache(ctx context.Context, uid int64, user *User) error {
	key := r.keyUser(uid)
	data, err := sonic.Marshal(user)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// GetUserCache 获取缓存的用户信息
func (r *RedisDao) GetUserCache(ctx context.Context, uid int64) (*User, error) {
	key := r.keyUser(uid)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var user User
	if err := sonic.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// DelUserCache 删除用户信息缓存
func (r *RedisDao) DelUserCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, r.keyUser(uid)).Err()
}

// DelUserCaches 批量删除用户信息缓存
func (r *RedisDao) DelUserCaches(ctx context.Context, uids []int64) error {
	if len(uids) == 0 {
		return nil
	}
	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = r.keyUser(uid)
	}
	return r.rdb.Del(ctx, keys...).Err()
}

// ---------- 用户配置缓存 ----------

// SetUserSettingsCache 缓存用户配置
func (r *RedisDao) SetUserSettingsCache(ctx context.Context, uid int64, cfg *UserConfig) error {
	key := r.keySettings(uid)
	data, err := sonic.Marshal(cfg)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// GetUserSettingsCache 获取缓存的用户配置
func (r *RedisDao) GetUserSettingsCache(ctx context.Context, uid int64) (*UserConfig, error) {
	key := r.keySettings(uid)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var cfg UserConfig
	if err := sonic.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// DelUserSettingsCache 删除用户配置缓存
func (r *RedisDao) DelUserSettingsCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, r.keySettings(uid)).Err()
}

// ---------- 封禁状态缓存 ----------

// SetBanCache 缓存封禁状态
func (r *RedisDao) SetBanCache(ctx context.Context, uid int64, banned bool, reason string) error {
	key := r.keyBan(uid)
	value := fmt.Sprintf("%d:%s", boolToInt(banned), reason)
	return r.rdb.Set(ctx, key, value, CacheTTL).Err()
}

// GetBanCache 获取缓存的封禁状态
func (r *RedisDao) GetBanCache(ctx context.Context, uid int64) (banned bool, reason string, err error) {
	key := r.keyBan(uid)
	data, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return false, "", err
	}
	// 简单解析格式 banned:reason
	parts := splitBanCache(data)
	if len(parts) >= 2 {
		banned = parts[0] == "1"
		reason = parts[1]
	}
	return banned, reason, nil
}

// DelBanCache 删除封禁状态缓存
func (r *RedisDao) DelBanCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, r.keyBan(uid)).Err()
}

// ---------- 设备缓存 ----------

// SetDevicesCache 缓存用户设备列表
func (r *RedisDao) SetDevicesCache(ctx context.Context, uid int64, devices []*UserDevice) error {
	key := r.keyDevices(uid)
	data, err := sonic.Marshal(devices)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// GetDevicesCache 获取缓存的设备列表
func (r *RedisDao) GetDevicesCache(ctx context.Context, uid int64) ([]*UserDevice, error) {
	key := r.keyDevices(uid)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var devices []*UserDevice
	if err := sonic.Unmarshal(data, &devices); err != nil {
		return nil, err
	}
	return devices, nil
}

// DelDevicesCache 删除设备缓存
func (r *RedisDao) DelDevicesCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, r.keyDevices(uid)).Err()
}

// ---------- 辅助函数 ----------

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func splitBanCache(data string) []string {
	// 找到第一个冒号的位置
	for i := 0; i < len(data); i++ {
		if data[i] == ':' {
			return []string{data[:i], data[i+1:]}
		}
	}
	return []string{data}
}

// SetUserStatusCache 设置用户在线状态缓存
func (r *RedisDao) SetUserStatusCache(ctx context.Context, uid int64, status string) error {
	key := fmt.Sprintf("user:status:%d", uid)
	return r.rdb.Set(ctx, key, status, 5*time.Minute).Err()
}

// GetUserStatusCache 获取用户在线状态缓存
func (r *RedisDao) GetUserStatusCache(ctx context.Context, uid int64) (string, error) {
	key := fmt.Sprintf("user:status:%d", uid)
	return r.rdb.Get(ctx, key).Result()
}

// IncrUserCounter 增加用户计数器（用于生成UID等）
func (r *RedisDao) IncrUserCounter(ctx context.Context, key string) (int64, error) {
	return r.rdb.Incr(ctx, key).Result()
}

// SetNX 设置键值（仅当键不存在时）
func (r *RedisDao) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.rdb.SetNX(ctx, key, value, expiration).Result()
}

// GetPushTokens 获取用户的所有推送令牌
func (r *RedisDao) GetPushTokens(ctx context.Context, uid int64) ([]string, error) {
	key := fmt.Sprintf("user:push_tokens:%d", uid)
	members, err := r.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return members, nil
}

// AddPushToken 添加推送令牌
func (r *RedisDao) AddPushToken(ctx context.Context, uid int64, token string) error {
	key := fmt.Sprintf("user:push_tokens:%d", uid)
	return r.rdb.SAdd(ctx, key, token).Err()
}

// RemovePushToken 移除推送令牌
func (r *RedisDao) RemovePushToken(ctx context.Context, uid int64, token string) error {
	key := fmt.Sprintf("user:push_tokens:%d", uid)
	return r.rdb.SRem(ctx, key, token).Err()
}

// SetUserOnline 设置用户在线状态
func (r *RedisDao) SetUserOnline(ctx context.Context, uid int64, platform string) error {
	key := fmt.Sprintf("user:online:%d", uid)
	return r.rdb.HSet(ctx, key, platform, strconv.FormatInt(time.Now().Unix(), 10)).Err()
}

// SetUserOffline 设置用户离线状态
func (r *RedisDao) SetUserOffline(ctx context.Context, uid int64, platform string) error {
	key := fmt.Sprintf("user:online:%d", uid)
	return r.rdb.HDel(ctx, key, platform).Err()
}

// IsUserOnline 检查用户是否在线
func (r *RedisDao) IsUserOnline(ctx context.Context, uid int64) (bool, error) {
	key := fmt.Sprintf("user:online:%d", uid)
	count, err := r.rdb.HLen(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
