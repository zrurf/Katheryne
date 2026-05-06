// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oss

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadPartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadPartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPartLogic {
	return &UploadPartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadPartLogic) UploadPart(req *types.UploadPartRequest) (resp *types.UploadPartResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
