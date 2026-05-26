package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFileInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFileInfoLogic {
	return &GetFileInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFileInfoLogic) GetFileInfo(req *types.GetDownloadURLRequest) (resp *types.GetFileInfoResponse, err error) {
	r, err := l.svcCtx.OssRpc.GetFileInfo(l.ctx, &ossclient.GetFileInfoReq{
		ObjectKey: req.ObjectKey,
		IndexId:   req.IndexId,
	})
	if err != nil {
		l.Errorf("GetFileInfo RPC failed: %v", err)
		return nil, err
	}
	return &types.GetFileInfoResponse{
		ObjectKey:    r.ObjectKey,
		Size:         r.Size,
		ContentType:  r.ContentType,
		Etag:         r.Etag,
		LastModified: r.LastModified,
		FileName:     r.FileName,
		IndexId:      r.IndexId,
	}, nil
}
