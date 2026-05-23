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

func (r *RedisDao) keyFriendship(uid int64) string {
	return fmt.Sprintf("social:friends:%d", uid)
}

func (r *RedisDao) keyGroupMembers(groupId int64) string {
	return fmt.Sprintf("social:group:members:%d", groupId)
}

func (r *RedisDao) keyGroupInfo(groupId int64) string {
	return fmt.Sprintf("social:group:info:%d", groupId)
}

func (r *RedisDao) keyBlacklist(uid int64) string {
	return fmt.Sprintf("social:blacklist:%d", uid)
}

func (r *RedisDao) DelFriendshipCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, r.keyFriendship(uid)).Err()
}

func (r *RedisDao) SetGroupMembersCache(ctx context.Context, groupId int64, memberUids []int64) error {
	key := r.keyGroupMembers(groupId)
	members := make([]interface{}, len(memberUids))
	for i, id := range memberUids {
		members[i] = strconv.FormatInt(id, 10)
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

func (r *RedisDao) IsGroupMemberCached(ctx context.Context, groupId, uid int64) (bool, error) {
	return r.rdb.SIsMember(ctx, r.keyGroupMembers(groupId), strconv.FormatInt(uid, 10)).Result()
}

func (r *RedisDao) DelGroupMembersCache(ctx context.Context, groupId int64) error {
	return r.rdb.Del(ctx, r.keyGroupMembers(groupId)).Err()
}

func (r *RedisDao) SetGroupInfoCache(ctx context.Context, groupId int64, fields map[string]interface{}) error {
	key := r.keyGroupInfo(groupId)
	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, CacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisDao) GetGroupInfoCache(ctx context.Context, groupId int64) (map[string]string, error) {
	return r.rdb.HGetAll(ctx, r.keyGroupInfo(groupId)).Result()
}

func (r *RedisDao) DelGroupInfoCache(ctx context.Context, groupId int64) error {
	return r.rdb.Del(ctx, r.keyGroupInfo(groupId)).Err()
}

func (r *RedisDao) DelBlacklistCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, r.keyBlacklist(uid)).Err()
}

func (r *RedisDao) keyUserOnline(uid int64) string {
	return fmt.Sprintf("social:user:online:%d", uid)
}

func (r *RedisDao) SetUserOnlineStatus(ctx context.Context, uid int64, status string) error {
	if status == "offline" {
		return r.rdb.Del(ctx, r.keyUserOnline(uid)).Err()
	}
	return r.rdb.Set(ctx, r.keyUserOnline(uid), status, CacheTTL).Err()
}

func (r *RedisDao) GetUserOnlineStatus(ctx context.Context, uid int64) (string, error) {
	status, err := r.rdb.Get(ctx, r.keyUserOnline(uid)).Result()
	if err == redis.Nil {
		return "offline", nil
	}
	if err != nil {
		return "", err
	}
	return status, nil
}

func (r *RedisDao) DelConvListCache(ctx context.Context, uid int64) error {
	return r.rdb.Del(ctx, fmt.Sprintf("conv:list:%d", uid)).Err()
}
