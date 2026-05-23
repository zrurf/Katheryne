package oss

import (
	"context"
	"net/url"

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
	// Return only the path (no scheme/host) so the App can assemble the
	// full URL using its own server host.
	proxyPath := "/api/v1/oss/file?" + url.Values{"key": {result.OssIndex}}.Encode()
	return &types.UploadResponse{
		FileName:  result.FileName,
		Size:      result.Size,
		Url:       proxyPath,
		OssIndex:  result.OssIndex,
		IndexId:   result.IndexId,
		ExpiresAt: result.ExpiresAt,
	}, nil
}
