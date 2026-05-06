package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return nil
}
