package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitiateUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitiateUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitiateUploadLogic {
	return &InitiateUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitiateUploadLogic) InitiateUpload(req *types.InitiateUploadRequest) (resp *types.InitiateUploadResponse, err error) {
	result, err := l.svcCtx.OssRpc.InitiateMultipartUpload(l.ctx, &ossclient.InitiateUploadReq{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		TotalSize:   req.TotalSize,
	})
	if err != nil {
		l.Errorf("InitiateMultipartUpload RPC failed: %v", err)
		return nil, err
	}
	return &types.InitiateUploadResponse{
		UploadID:  result.UploadId,
		Bucket:    result.Bucket,
		ObjectKey: result.ObjectKey,
	}, nil
}
