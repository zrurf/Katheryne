package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSettingsLogic {
	return &GetUserSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserSettingsLogic) GetUserSettings(in *user.GetUserSettingsReq) (*user.GetUserSettingsResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetUserSettingsResp{}, nil
}
