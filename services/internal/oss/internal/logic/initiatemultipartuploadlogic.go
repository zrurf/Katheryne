package logic

import (
	"context"
	"time"

	"oss/internal/dao"
	"oss/internal/svc"
	"oss/oss"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type InitiateMultipartUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitiateMultipartUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitiateMultipartUploadLogic {
	return &InitiateMultipartUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化分片上传，返回 UploadId
func (l *InitiateMultipartUploadLogic) InitiateMultipartUpload(in *oss.InitiateUploadReq) (*oss.InitiateUploadResp, error) {
	if in.FileName == "" {
		return nil, nil
	}

	uploadID := uuid.New().String()
	objectKey := dao.BuildObjectKey(uploadID, in.FileName)

	// 调用 RustFS 初始化分片上传
	realUploadID, err := l.svcCtx.Storage.InitiateMultipartUpload(l.ctx, objectKey)
	if err != nil {
		l.Errorf("InitiateMultipartUpload error: %v", err)
		return nil, err
	}

	// 缓存上传元数据
	meta := &dao.UploadMeta{
		UploadID:    realUploadID,
		Bucket:      l.svcCtx.Storage.GetBucket(),
		ObjectKey:   objectKey,
		FileName:    in.FileName,
		ContentType: in.ContentType,
		TotalSize:   in.TotalSize,
		CreatedAt:   time.Now().Unix(),
	}
	if err := l.svcCtx.RedisDao.SetUploadMeta(l.ctx, meta); err != nil {
		l.Errorf("SetUploadMeta error: %v", err)
	}

	return &oss.InitiateUploadResp{
		UploadId:  realUploadID,
		Bucket:    meta.Bucket,
		ObjectKey: objectKey,
	}, nil
}
