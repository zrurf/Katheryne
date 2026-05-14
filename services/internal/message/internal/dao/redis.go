package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CacheTTL = 10 * time.Minute
)

type RedisDao struct {
	rdb *redis.Client
}

func NewRedisDao(rdb *redis.Client) *RedisDao {
	return &RedisDao{rdb: rdb}
}

func (r *RedisDao) keyUnreadCount(uid, convId int64) string {
	return fmt.Sprintf("msg:unread:%d:%d", uid, convId)
}

func (r *RedisDao) keyLastMsg(convId int64) string {
	return fmt.Sprintf("msg:last:%d", convId)
}

func (r *RedisDao) keyReadMembers(convId, msgId int64) string {
	return fmt.Sprintf("msg:read:%d:%d", convId, msgId)
}

func (r *RedisDao) GetUnreadCache(ctx context.Context, uid, convId int64) (int64, error) {
	val, err := r.rdb.Get(ctx, r.keyUnreadCount(uid, convId)).Result()
	if err == redis.Nil {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	v, _ := strconv.ParseInt(val, 10, 64)
	return v, nil
}

func (r *RedisDao) SetUnreadCache(ctx context.Context, uid, convId, count int64) error {
	return r.rdb.Set(ctx, r.keyUnreadCount(uid, convId), count, CacheTTL).Err()
}

func (r *RedisDao) DelUnreadCache(ctx context.Context, uid, convId int64) error {
	return r.rdb.Del(ctx, r.keyUnreadCount(uid, convId)).Err()
}

func (r *RedisDao) DelUnreadCaches(ctx context.Context, uid int64, convIds []int64) error {
	if len(convIds) == 0 {
		return nil
	}
	keys := make([]string, len(convIds))
	for i, id := range convIds {
		keys[i] = r.keyUnreadCount(uid, id)
	}
	return r.rdb.Del(ctx, keys...).Err()
}

func (r *RedisDao) SetLastMsgCache(ctx context.Context, convId int64, msgId int64, snippet string, sender int64, createdAt int64) error {
	key := r.keyLastMsg(convId)
	data := map[string]interface{}{
		"msg_id":     msgId,
		"snippet":    snippet,
		"sender":     sender,
		"created_at": createdAt,
	}
	return r.rdb.HSet(ctx, key, data).Err()
}

func (r *RedisDao) GetLastMsgCache(ctx context.Context, convId int64) (map[string]string, error) {
	return r.rdb.HGetAll(ctx, r.keyLastMsg(convId)).Result()
}

func (r *RedisDao) DelLastMsgCache(ctx context.Context, convId int64) error {
	return r.rdb.Del(ctx, r.keyLastMsg(convId)).Err()
}

func (r *RedisDao) AddReadMemberCache(ctx context.Context, convId, msgId int64, uid int64, readAt int64) error {
	key := r.keyReadMembers(convId, msgId)
	return r.rdb.HSet(ctx, key, strconv.FormatInt(uid, 10), readAt).Err()
}

func (r *RedisDao) GetReadMemberCache(ctx context.Context, convId, msgId int64) (map[string]string, error) {
	return r.rdb.HGetAll(ctx, r.keyReadMembers(convId, msgId)).Result()
}

func (r *RedisDao) ExpireReadMemberCache(ctx context.Context, convId, msgId int64) error {
	return r.rdb.Expire(ctx, r.keyReadMembers(convId, msgId), CacheTTL).Err()
}
