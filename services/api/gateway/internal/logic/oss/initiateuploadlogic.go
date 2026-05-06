// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
