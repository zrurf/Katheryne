package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	UploadCachePrefix = "oss:upload:"
	PartCachePrefix   = "oss:part:"
	URLCachePrefix    = "oss:url:"
	CacheTTL          = 24 * time.Hour
)

// UploadMeta 上传元数据
type UploadMeta struct {
	UploadID    string `json:"upload_id"`
	Bucket      string `json:"bucket"`
	ObjectKey   string `json:"object_key"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	TotalSize   int64  `json:"total_size"`
	CreatedAt   int64  `json:"created_at"`
}

// PartMeta 分片元数据
type PartMeta struct {
	PartNumber int32  `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
	UploadedAt int64  `json:"uploaded_at"`
}

// RedisDao Redis 缓存层
type RedisDao struct {
	rdb *redis.Client
}

func NewRedisDao(rdb *redis.Client) *RedisDao {
	return &RedisDao{rdb: rdb}
}

func (r *RedisDao) keyUpload(uploadID string) string {
	return fmt.Sprintf("%s%s", UploadCachePrefix, uploadID)
}

func (r *RedisDao) keyParts(uploadID string) string {
	return fmt.Sprintf("%s%s", PartCachePrefix, uploadID)
}

func (r *RedisDao) keyURL(objectKey string) string {
	return fmt.Sprintf("%s%s", URLCachePrefix, objectKey)
}

// SetUploadMeta 缓存上传元数据
func (r *RedisDao) SetUploadMeta(ctx context.Context, meta *UploadMeta) error {
	key := r.keyUpload(meta.UploadID)
	data, err := sonic.Marshal(meta)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// GetUploadMeta 获取上传元数据
func (r *RedisDao) GetUploadMeta(ctx context.Context, uploadID string) (*UploadMeta, error) {
	key := r.keyUpload(uploadID)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var meta UploadMeta
	if err := sonic.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// DelUploadMeta 删除上传元数据
func (r *RedisDao) DelUploadMeta(ctx context.Context, uploadID string) error {
	return r.rdb.Del(ctx, r.keyUpload(uploadID)).Err()
}

// AddPart 记录已上传分片
func (r *RedisDao) AddPart(ctx context.Context, uploadID string, part *PartMeta) error {
	key := r.keyParts(uploadID)
	data, err := sonic.Marshal(part)
	if err != nil {
		return err
	}
	return r.rdb.HSet(ctx, key, fmt.Sprintf("%d", part.PartNumber), data).Err()
}

// GetParts 获取所有已上传分片
func (r *RedisDao) GetParts(ctx context.Context, uploadID string) ([]*PartMeta, error) {
	key := r.keyParts(uploadID)
	vals, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var parts []*PartMeta
	for _, v := range vals {
		var part PartMeta
		if err := sonic.Unmarshal([]byte(v), &part); err != nil {
			continue
		}
		parts = append(parts, &part)
	}
	return parts, nil
}

// DelParts 删除分片记录
func (r *RedisDao) DelParts(ctx context.Context, uploadID string) error {
	return r.rdb.Del(ctx, r.keyParts(uploadID)).Err()
}

// SetURLCache 缓存签名 URL
func (r *RedisDao) SetURLCache(ctx context.Context, objectKey, urlStr string, expiresAt int64) error {
	key := r.keyURL(objectKey)
	value := fmt.Sprintf("%s|%d", urlStr, expiresAt)
	return r.rdb.Set(ctx, key, value, CacheTTL).Err()
}

// GetURLCache 获取缓存的 URL
func (r *RedisDao) GetURLCache(ctx context.Context, objectKey string) (string, int64, error) {
	key := r.keyURL(objectKey)
	data, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", 0, err
	}
	var urlStr string
	var expiresAt int64
	fmt.Sscanf(data, "%s|%d", &urlStr, &expiresAt)
	return urlStr, expiresAt, nil
}

// DelURLCache 删除 URL 缓存
func (r *RedisDao) DelURLCache(ctx context.Context, objectKey string) error {
	return r.rdb.Del(ctx, r.keyURL(objectKey)).Err()
}
