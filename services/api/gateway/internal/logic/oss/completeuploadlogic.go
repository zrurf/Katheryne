package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompleteUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompleteUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteUploadLogic {
	return &CompleteUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompleteUploadLogic) CompleteUpload(req *types.CompleteUploadRequest) (resp *types.UploadResponse, err error) {
	parts := make([]*ossclient.PartInfo, len(req.Parts))
	for i, p := range req.Parts {
		parts[i] = &ossclient.PartInfo{
			PartNumber: int32(p.PartNumber),
			Etag:       p.ETag,
		}
	}
	result, err := l.svcCtx.OssRpc.CompleteMultipartUpload(l.ctx, &ossclient.CompleteUploadReq{
		UploadId: req.UploadID,
		Parts:    parts,
	})
	if err != nil {
		l.Errorf("CompleteMultipartUpload RPC failed: %v", err)
		return nil, err
	}
	return &types.UploadResponse{
		FileName:  result.OssIndex,
		Size:      result.Size,
		Url:       result.Url,
		OssIndex:  result.OssIndex,
		ExpiresAt: result.ExpiresAt,
	}, nil
}