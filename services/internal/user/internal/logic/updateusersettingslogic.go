package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserSettingsLogic {
	return &UpdateUserSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserSettingsLogic) UpdateUserSettings(in *user.UpdateUserSettingsReq) (*user.UpdateUserSettingsResp, error) {
	// todo: add your logic here and delete this line

	return &user.UpdateUserSettingsResp{}, nil
}
