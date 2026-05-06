package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddPushTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddPushTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPushTokenLogic {
	return &AddPushTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddPushTokenLogic) AddPushToken(in *user.AddPushTokenReq) (*user.AddPushTokenResp, error) {
	// todo: add your logic here and delete this line

	return &user.AddPushTokenResp{}, nil
}
