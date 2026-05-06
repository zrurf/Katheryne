package logic

import (
	"context"

	"oss/internal/svc"
	"oss/oss"

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
	// todo: add your logic here and delete this line

	return &oss.UploadResp{}, nil
}
