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

	// 确保 bucket 存在
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region})
		if err != nil {
			return nil, err
		}
	}

	return &StorageDao{
		core:   core,
		client: client,
		bucket: bucket,
	}, nil
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

// StatObject 获取对象信息（大小等）
func (d *StorageDao) StatObject(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	return d.client.StatObject(ctx, d.bucket, objectKey, minio.StatObjectOptions{})
}

// RemoveObject 删除对象
func (d *StorageDao) RemoveObject(ctx context.Context, objectKey string) error {
	return d.client.RemoveObject(ctx, d.bucket, objectKey, minio.RemoveObjectOptions{})
}

// BuildObjectKey 构建对象存储路径：uploads/{uploadID}/{fileName}
func BuildObjectKey(uploadID, fileName string) string {
	return path.Join("uploads", uploadID, fileName)
}

// BuildURL 生成访问 URL
func (d *StorageDao) BuildURL(ctx context.Context, objectKey string) (string, int64, error) {
	// 默认 URL 有效期 7 天
	expiry := 7 * 24 * time.Hour
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
