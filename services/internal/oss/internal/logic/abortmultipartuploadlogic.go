package logic

import (
	"context"

	"oss/internal/svc"
	"oss/oss"

	"github.com/zeromicro/go-zero/core/logx"
)

type AbortMultipartUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAbortMultipartUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AbortMultipartUploadLogic {
	return &AbortMultipartUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消上传
func (l *AbortMultipartUploadLogic) AbortMultipartUpload(in *oss.AbortUploadReq) (*oss.AbortUploadResp, error) {
	// todo: add your logic here and delete this line

	return &oss.AbortUploadResp{}, nil
}
