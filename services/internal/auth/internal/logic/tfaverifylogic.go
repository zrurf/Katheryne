package logic

import (
	"context"

	"auth/auth"
	"auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TFAVerifyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTFAVerifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TFAVerifyLogic {
	return &TFAVerifyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TFAVerifyLogic) TFAVerify(in *auth.TFAVerifyReq) (*auth.TFAVerifyResp, error) {
	// todo: add your logic here and delete this line

	return &auth.TFAVerifyResp{}, nil
}
