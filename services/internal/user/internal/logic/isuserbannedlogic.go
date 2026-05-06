package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsUserBannedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsUserBannedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsUserBannedLogic {
	return &IsUserBannedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsUserBannedLogic) IsUserBanned(in *user.IsUserBannedReq) (*user.IsUserBannedResp, error) {
	// todo: add your logic here and delete this line

	return &user.IsUserBannedResp{}, nil
}
