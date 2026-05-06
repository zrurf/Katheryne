// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"auth/auth"
	"context"
	"encoding/base64"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type RegisterFinalizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterFinalizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterFinalizeLogic {
	return &RegisterFinalizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterFinalizeLogic) RegisterFinalize(req *types.RegisterFinalizeRequest) (resp *types.RegisterFinalizeResponse, err error) {
	registrationRecord, err := base64.URLEncoding.DecodeString(req.RegistrationRecord)
	if err != nil {
		l.Errorf("RegistrationRecord Base64 decoding failed: %v", err)
		return nil, errors.New(104, "Request Error")
	}
	result, err := l.svcCtx.AuthRpc.RegisterFinalize(l.ctx, &auth.RegisterFinalizeReq{
		Phone:              req.Phone,
		Name:               req.Name,
		RegistrationRecord: registrationRecord,
	})
	if err != nil {
		l.Errorf("RegisterFinalize RPC failed: %v", err)
		return nil, err
	}
	return &types.RegisterFinalizeResponse{
		Result: result.Result,
		Reason: result.Reason,
	}, nil
}
