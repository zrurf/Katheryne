package logic

import (
	"context"
	"sort"

	"oss/internal/svc"
	"oss/oss"

	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type CompleteMultipartUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompleteMultipartUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteMultipartUploadLogic {
	return &CompleteMultipartUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 完成分片上传（合并）
func (l *CompleteMultipartUploadLogic) CompleteMultipartUpload(in *oss.CompleteUploadReq) (*oss.UploadResp, error) {
	if in.UploadId == "" || in.ObjectKey == "" {
		return nil, nil
	}

	// 从缓存获取已上传分片信息（作为校验和兜底）
	cachedParts, err := l.svcCtx.RedisDao.GetParts(l.ctx, in.UploadId)
	if err != nil {
		l.Infof("GetParts from cache error: %v", err)
	}

	// 构建 CompletePart 列表，优先使用请求中的 parts
	var parts []minio.CompletePart
	if len(in.Parts) > 0 {
		for _, p := range in.Parts {
			parts = append(parts, minio.CompletePart{
				PartNumber: int(p.PartNumber),
				ETag:       p.Etag,
			})
		}
	} else if len(cachedParts) > 0 {
		for _, p := range cachedParts {
			parts = append(parts, minio.CompletePart{
				PartNumber: int(p.PartNumber),
				ETag:       p.ETag,
			})
		}
	}

	// 按 PartNumber 排序（S3 要求）
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	// 调用 RustFS 完成合并
	info, err := l.svcCtx.Storage.CompleteMultipartUpload(l.ctx, in.UploadId, in.ObjectKey, parts)
	if err != nil {
		l.Errorf("CompleteMultipartUpload error: %v", err)
		return nil, err
	}

	// 生成访问 URL
	urlStr, expiresAt, err := l.svcCtx.Storage.BuildURL(l.ctx, in.ObjectKey)
	if err != nil {
		l.Errorf("BuildURL error: %v", err)
		// 即使 URL 生成失败，也返回合并成功的结果
		urlStr = l.svcCtx.Storage.ComposePublicURL(in.ObjectKey)
		expiresAt = 0
	}

	// 缓存 URL
	go func() {
		_ = l.svcCtx.RedisDao.SetURLCache(context.Background(), in.ObjectKey, urlStr, expiresAt)
	}()

	// 清理上传元数据和分片缓存
	go func() {
		_ = l.svcCtx.RedisDao.DelUploadMeta(context.Background(), in.UploadId)
		_ = l.svcCtx.RedisDao.DelParts(context.Background(), in.UploadId)
	}()

	return &oss.UploadResp{
		Url:       urlStr,
		Size:      info.Size,
		OssIndex:  in.ObjectKey,
		ExpiresAt: expiresAt,
	}, nil
}
