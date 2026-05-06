package logic

import (
	"context"

	"oss/internal/svc"
	"oss/oss"

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
	// todo: add your logic here and delete this line

	return &oss.InitiateUploadResp{}, nil
}
