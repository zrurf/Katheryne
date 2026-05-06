package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemovePushTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemovePushTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePushTokenLogic {
	return &RemovePushTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemovePushTokenLogic) RemovePushToken(in *user.RemovePushTokenReq) (*user.RemovePushTokenResp, error) {
	// todo: add your logic here and delete this line

	return &user.RemovePushTokenResp{}, nil
}
