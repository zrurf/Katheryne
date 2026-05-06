// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
