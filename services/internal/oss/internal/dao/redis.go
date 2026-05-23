package dao

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	UploadCachePrefix = "oss:upload:"
	PartCachePrefix   = "oss:part:"
	URLCachePrefix    = "oss:url:"
	IndexPrefix       = "oss:index:"
	IndexByHashPrefix = "oss:index_by_hash:"
	CacheTTL          = 24 * time.Hour
	IndexTTL          = 30 * 24 * time.Hour // 30 days for index entries
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
	// URL 和 expiresAt 以最后一个 | 分隔
	idx := strings.LastIndex(data, "|")
	if idx < 0 {
		return data, 0, nil
	}
	urlStr := data[:idx]
	expiresAt, _ := strconv.ParseInt(data[idx+1:], 10, 64)
	return urlStr, expiresAt, nil
}

// DelURLCache 删除 URL 缓存
func (r *RedisDao) DelURLCache(ctx context.Context, objectKey string) error {
	return r.rdb.Del(ctx, r.keyURL(objectKey)).Err()
}

// ========== File Index (permanent file metadata, keyed by index_id) ==========

// FileIndex 文件索引（持久化的文件元数据，与内容分离存储）
type FileIndex struct {
	IndexID     string `json:"index_id"`     // UUID, the permanent reference
	FileName    string `json:"file_name"`    // 原始文件名
	ContentType string `json:"content_type"` // MIME 类型
	ObjectKey   string `json:"object_key"`   // 内容哈希键（指向存储对象）
	Size        int64  `json:"size"`         // 文件大小
	CreatedAt   int64  `json:"created_at"`   // 创建时间戳
}

func (r *RedisDao) keyIndex(indexID string) string {
	return fmt.Sprintf("%s%s", IndexPrefix, indexID)
}

func (r *RedisDao) keyIndexByHash(objectKey string) string {
	return fmt.Sprintf("%s%s", IndexByHashPrefix, objectKey)
}

// SetFileIndex 创建文件索引，同时注册到 hash→index_ids 集合
func (r *RedisDao) SetFileIndex(ctx context.Context, idx *FileIndex) error {
	key := r.keyIndex(idx.IndexID)
	data, err := sonic.Marshal(idx)
	if err != nil {
		return err
	}
	pipe := r.rdb.Pipeline()
	pipe.Set(ctx, key, data, IndexTTL)
	pipe.SAdd(ctx, r.keyIndexByHash(idx.ObjectKey), idx.IndexID)
	_, err = pipe.Exec(ctx)
	return err
}

// GetFileIndex 通过 index_id 获取文件索引
func (r *RedisDao) GetFileIndex(ctx context.Context, indexID string) (*FileIndex, error) {
	data, err := r.rdb.Get(ctx, r.keyIndex(indexID)).Bytes()
	if err != nil {
		return nil, err
	}
	var idx FileIndex
	if err := sonic.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// GetIndexesByHash 通过内容哈希获取所有指向该内容的 index_ids
func (r *RedisDao) GetIndexesByHash(ctx context.Context, objectKey string) ([]string, error) {
	return r.rdb.SMembers(ctx, r.keyIndexByHash(objectKey)).Result()
}

// DelFileIndex 删除文件索引（同时从 hash 集合中移除）
func (r *RedisDao) DelFileIndex(ctx context.Context, indexID string) error {
	idx, err := r.GetFileIndex(ctx, indexID)
	if err != nil {
		return err
	}
	pipe := r.rdb.Pipeline()
	pipe.Del(ctx, r.keyIndex(indexID))
	pipe.SRem(ctx, r.keyIndexByHash(idx.ObjectKey), indexID)
	_, err = pipe.Exec(ctx)
	return err
}
