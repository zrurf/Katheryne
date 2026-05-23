package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFileLogic {
	return &DeleteFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFileLogic) DeleteFile(req *types.DeleteFileRequest) (resp *types.DeleteFileResponse, err error) {
	_, err = l.svcCtx.OssRpc.DeleteFile(l.ctx, &ossclient.DeleteFileReq{
		ObjectKey: req.ObjectKey,
		IndexId:   req.IndexId,
	})
	if err != nil {
		l.Errorf("DeleteFile RPC failed: %v", err)
		return nil, err
	}
	return &types.DeleteFileResponse{Deleted: true}, nil
}