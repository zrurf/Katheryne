package logic

import (
	"bytes"
	"context"
	"time"

	"oss/internal/dao"
	"oss/internal/svc"
	"oss/oss"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadPartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadPartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPartLogic {
	return &UploadPartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 上传分片（流式）
func (l *UploadPartLogic) UploadPart(stream oss.OSS_UploadPartServer) error {
	var uploadID string
	var objectKey string
	var partNumber int32

	for {
		req, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			l.Errorf("stream recv error: %v", err)
			return err
		}

		// 第一次接收时记录 uploadID 和 partNumber
		if uploadID == "" {
			uploadID = req.UploadId
			partNumber = req.PartNumber

			// 从缓存获取 objectKey
			meta, err := l.svcCtx.RedisDao.GetUploadMeta(l.ctx, uploadID)
			if err != nil || meta == nil {
				l.Errorf("GetUploadMeta error: %v", err)
				return err
			}
			objectKey = meta.ObjectKey
		}

		// 上传分片到 RustFS
		data := req.Data
		if len(data) == 0 {
			continue
		}

		etag, err := l.svcCtx.Storage.UploadPart(
			l.ctx,
			uploadID,
			int(partNumber),
			objectKey,
			bytes.NewReader(data),
			int64(len(data)),
		)
		if err != nil {
			l.Errorf("UploadPart error: %v", err)
			return err
		}

		// 缓存分片信息
		partMeta := &dao.PartMeta{
			PartNumber: partNumber,
			ETag:       etag,
			Size:       int64(len(data)),
			UploadedAt: time.Now().Unix(),
		}
		if err := l.svcCtx.RedisDao.AddPart(l.ctx, uploadID, partMeta); err != nil {
			l.Errorf("AddPart cache error: %v", err)
		}

		// 返回响应
		if err := stream.SendAndClose(&oss.UploadPartResp{
			Etag:       etag,
			PartNumber: partNumber,
		}); err != nil {
			l.Errorf("SendAndClose error: %v", err)
			return err
		}
		break
	}

	return nil
}
