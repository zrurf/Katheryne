package oss

import (
	"context"
	"net/url"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDownloadURLLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDownloadURLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDownloadURLLogic {
	return &GetDownloadURLLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDownloadURLLogic) GetDownloadURL(req *types.GetDownloadURLRequest) (resp *types.GetDownloadURLResponse, err error) {
	r, err := l.svcCtx.OssRpc.GetDownloadURL(l.ctx, &ossclient.GetDownloadURLReq{
		ObjectKey:  req.ObjectKey,
		IndexId:    req.IndexId,
		ExpireSecs: req.ExpireSecs,
	})
	if err != nil {
		l.Errorf("GetDownloadURL RPC failed: %v", err)
		return nil, err
	}
	// Return only the path (no scheme/host) so the App can assemble the
	// full URL using its own server host.
	resultURL := r.Url
	if req.ObjectKey != "" {
		proxyPath := "/api/v1/oss/file?" + url.Values{"key": {req.ObjectKey}}.Encode()
		resultURL = proxyPath
	}
	return &types.GetDownloadURLResponse{
		Url:       resultURL,
		ExpiresAt: r.ExpiresAt,
	}, nil
}
