package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ConvCachePrefix       = "conv:"
	ConvMembersCachePrefix = "conv:members:"
	ConvListCachePrefix   = "conv:list:"
	UnreadCachePrefix     = "unread:"
	CacheTTL              = 10 * time.Minute
)

type RedisDao struct {
	rdb *redis.Client
}

func NewRedisDao(rdb *redis.Client) *RedisDao {
	return &RedisDao{rdb: rdb}
}

func (r *RedisDao) keyConv(convId int64) string {
	return fmt.Sprintf("%s%d", ConvCachePrefix, convId)
}

func (r *RedisDao) keyConvMembers(convId int64) string {
	return fmt.Sprintf("%s%d", ConvMembersCachePrefix, convId)
}

func (r *RedisDao) keyConvList(uid int64) string {
	return fmt.Sprintf("%s%d", ConvListCachePrefix, uid)
}

func (r *RedisDao) keyUnread(uid, convId int64) string {
	return fmt.Sprintf("%s%d:%d", UnreadCachePrefix, uid, convId)
}

func (r *RedisDao) DelConvCache(ctx context.Context, convId int64) error {
	return r.rdb.Del(ctx, r.keyConv(convId)).Err()
}

func (r *RedisDao) DelConvMembersCache(ctx context.Context, convId int64) error {
	return r.rdb.Del(ctx, r.keyConvMembers(convId)).Err()
}

func (r *RedisDao) DelConvListCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, r.keyConvList(uid)).Err()
}

func (r *RedisDao) DelConvListCaches(ctx context.Context, uids []int64) error {
	pipe := r.rdb.Pipeline()
	for _, uid := range uids {
		pipe.Del(ctx, r.keyConvList(uid))
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisDao) SetConvMembersCache(ctx context.Context, convId int64, uids []int64) error {
	key := r.keyConvMembers(convId)
	members := make([]interface{}, len(uids))
	for i, uid := range uids {
		members[i] = strconv.FormatInt(uid, 10)
	}
	pipe := r.rdb.Pipeline()
	pipe.Del(ctx, key)
	if len(members) > 0 {
		pipe.SAdd(ctx, key, members...)
	}
	pipe.Expire(ctx, key, CacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisDao) GetConvMembersCache(ctx context.Context, convId int64) ([]int64, error) {
	key := r.keyConvMembers(convId)
	members, err := r.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, redis.Nil
	}
	uids := make([]int64, 0, len(members))
	for _, m := range members {
		uid, err := strconv.ParseInt(m, 10, 64)
		if err != nil {
			continue
		}
		uids = append(uids, uid)
	}
	return uids, nil
}

func (r *RedisDao) SetUnreadCache(ctx context.Context, uid, convId, count int64) error {
	key := r.keyUnread(uid, convId)
	return r.rdb.Set(ctx, key, count, CacheTTL).Err()
}

func (r *RedisDao) GetUnreadCache(ctx context.Context, uid, convId int64) (int64, error) {
	key := r.keyUnread(uid, convId)
	v, err := r.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return -1, nil
	}
	return v, err
}

func (r *RedisDao) DelUnreadCache(ctx context.Context, uid, convId int64) error {
	return r.rdb.Del(ctx, r.keyUnread(uid, convId)).Err()
}

func (r *RedisDao) IncrUnreadCache(ctx context.Context, uid, convId int64) error {
	key := r.keyUnread(uid, convId)
	pipe := r.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, CacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}
