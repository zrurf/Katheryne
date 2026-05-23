package dao

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageDao 对象存储访问层（基于 S3 兼容 API）
type StorageDao struct {
	core   *minio.Core
	client *minio.Client
	bucket string
}

// NewStorageDao 创建存储 DAO
func NewStorageDao(endpoint, accessKey, secretKey, bucket, region string, useSSL bool) (*StorageDao, error) {
	endpoint = cleanEndpoint(endpoint)

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}

	core, err := minio.NewCore(endpoint, opts)
	if err != nil {
		return nil, err
	}

	d := &StorageDao{
		core:   core,
		client: client,
		bucket: bucket,
	}

	// Ensure bucket exists — non‑fatal on startup.
	// The bucket should already exist from initial deployment.
	// If this is a transient DNS/network issue, operations will still
	// work once connectivity is restored.
	if err := d.ensureBucket(); err != nil {
		// Log would go here if we had a logger; caller handles it.
		return d, err
	}

	return d, nil
}

// ensureBucket checks bucket existence and creates it if missing.
func (d *StorageDao) ensureBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := d.client.BucketExists(ctx, d.bucket)
	if err != nil {
		return fmt.Errorf("bucket check failed: %w", err)
	}
	if !exists {
		if err := d.client.MakeBucket(ctx, d.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket failed: %w", err)
		}
	}
	return nil
}

// InitiateMultipartUpload 初始化分片上传，返回 uploadID
func (d *StorageDao) InitiateMultipartUpload(ctx context.Context, objectKey string) (string, error) {
	uploadID, err := d.core.NewMultipartUpload(ctx, d.bucket, objectKey, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	return uploadID, nil
}

// UploadPart 上传单个分片
func (d *StorageDao) UploadPart(ctx context.Context, uploadID string, partNumber int, objectKey string, data io.Reader, size int64) (string, error) {
	part, err := d.core.PutObjectPart(ctx, d.bucket, objectKey, uploadID, partNumber, data, size, minio.PutObjectPartOptions{})
	if err != nil {
		return "", err
	}
	return part.ETag, nil
}

// CompleteMultipartUpload 完成分片上传（合并）
func (d *StorageDao) CompleteMultipartUpload(ctx context.Context, uploadID, objectKey string, parts []minio.CompletePart) (minio.UploadInfo, error) {
	info, err := d.core.CompleteMultipartUpload(ctx, d.bucket, objectKey, uploadID, parts, minio.PutObjectOptions{})
	if err != nil {
		return minio.UploadInfo{}, err
	}
	return info, nil
}

// AbortMultipartUpload 取消分片上传
func (d *StorageDao) AbortMultipartUpload(ctx context.Context, uploadID, objectKey string) error {
	err := d.core.AbortMultipartUpload(ctx, d.bucket, objectKey, uploadID)
	return err
}

// PresignedGetObject 生成带签名的访问 URL
func (d *StorageDao) PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (*url.URL, error) {
	return d.client.PresignedGetObject(ctx, d.bucket, objectKey, expiry, nil)
}

// PutObjectSimple 简单上传（非分片，直接存储整个对象）
func (d *StorageDao) PutObjectSimple(ctx context.Context, objectKey, contentType string, data io.Reader, size int64) (minio.UploadInfo, error) {
	return d.client.PutObject(ctx, d.bucket, objectKey, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

// CopyObject copies an object within the same bucket (used to move temp upload to hash-based key).
func (d *StorageDao) CopyObject(ctx context.Context, srcKey, dstKey string, contentType string) (minio.UploadInfo, error) {
	src := minio.CopySrcOptions{Bucket: d.bucket, Object: srcKey}
	dst := minio.CopyDestOptions{Bucket: d.bucket, Object: dstKey}
	info, err := d.client.CopyObject(ctx, dst, src)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	return minio.UploadInfo{
		Key:  dstKey,
		Size: info.Size,
	}, nil
}

// GetObject returns a reader for streaming an object's content.
func (d *StorageDao) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	return d.client.GetObject(ctx, d.bucket, objectKey, minio.GetObjectOptions{})
}

// StatObject 获取对象信息（大小等）
func (d *StorageDao) StatObject(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	return d.client.StatObject(ctx, d.bucket, objectKey, minio.StatObjectOptions{})
}

// RemoveObject 删除对象
func (d *StorageDao) RemoveObject(ctx context.Context, objectKey string) error {
	return d.client.RemoveObject(ctx, d.bucket, objectKey, minio.RemoveObjectOptions{})
}

// BuildObjectKey 基于文件 Hash 构建对象存储路径：uploads/{hash[:2]}/{hash}
// 相同文件（相同 Hash）永远只存储一次，URL 不包含原始文件名。
func BuildObjectKey(fileHash string) string {
	if len(fileHash) < 4 {
		return path.Join("uploads", "unknown", fileHash)
	}
	return path.Join("uploads", fileHash[:2], fileHash)
}

// BuildURL 生成访问 URL（默认 7 天有效期）
func (d *StorageDao) BuildURL(ctx context.Context, objectKey string) (string, int64, error) {
	return d.BuildURLWithExpiry(ctx, objectKey, 7*24*time.Hour)
}

// BuildURLWithExpiry 生成带自定义有效期的访问 URL
func (d *StorageDao) BuildURLWithExpiry(ctx context.Context, objectKey string, expiry time.Duration) (string, int64, error) {
	u, err := d.PresignedGetObject(ctx, objectKey, expiry)
	if err != nil {
		return "", 0, err
	}
	expiresAt := time.Now().Add(expiry).Unix()
	return u.String(), expiresAt, nil
}

// GetBucket 返回 bucket 名称
func (d *StorageDao) GetBucket() string {
	return d.bucket
}

// GetEndpoint 返回 endpoint
func (d *StorageDao) GetEndpoint() string {
	return d.client.EndpointURL().String()
}

// ComposePublicURL 组合公开访问 URL（如果 RustFS 配置为公开访问）
func (d *StorageDao) ComposePublicURL(objectKey string) string {
	endpoint := d.GetEndpoint()
	u, _ := url.Parse(endpoint)
	u.Path = path.Join(d.bucket, objectKey)
	return u.String()
}

// ValidateUploadID 简单校验 uploadID 格式
func cleanEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}
	endpoint = strings.TrimRight(endpoint, "/")
	return endpoint
}

func ValidateUploadID(uploadID string) error {
	if uploadID == "" {
		return fmt.Errorf("upload_id is empty")
	}
	return nil
}
