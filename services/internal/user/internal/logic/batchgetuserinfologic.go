package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserInfoLogic {
	return &BatchGetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchGetUserInfoLogic) BatchGetUserInfo(in *user.BatchGetUserInfoReq) (*user.BatchGetUserInfoResp, error) {
	// todo: add your logic here and delete this line

	return &user.BatchGetUserInfoResp{}, nil
}
