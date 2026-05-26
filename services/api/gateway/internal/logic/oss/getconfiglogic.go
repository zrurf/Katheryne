package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigLogic) GetConfig() (resp *types.OssConfigResponse, err error) {
	maxFileSize := l.svcCtx.Config.MaxFileSize
	if maxFileSize <= 0 {
		maxFileSize = 104857600 // 100 MB default
	}
	return &types.OssConfigResponse{
		MaxFileSize: maxFileSize,
		ChunkSize:   5 * 1024 * 1024, // 5 MB chunks
		MaxParts:    100,
	}, nil
}
