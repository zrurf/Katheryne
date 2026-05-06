package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByUIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByUIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByUIDLogic {
	return &GetUserByUIDLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByUIDLogic) GetUserByUID(in *user.GetUserByUIDReq) (*user.GetUserResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetUserResp{}, nil
}
