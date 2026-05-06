package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLastLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLastLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLastLoginLogic {
	return &UpdateLastLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateLastLoginLogic) UpdateLastLogin(in *user.UpdateLastLoginReq) (*user.UpdateLastLoginResp, error) {
	// todo: add your logic here and delete this line

	return &user.UpdateLastLoginResp{}, nil
}
