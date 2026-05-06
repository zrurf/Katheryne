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

type RegisterInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterInitLogic {
	return &RegisterInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterInitLogic) RegisterInit(req *types.RegisterInitRequest) (resp *types.RegisterInitResponse, err error) {
	registrationRequest, err := base64.URLEncoding.DecodeString(req.RegistrationRequest)
	if err != nil {
		l.Errorf("RegistrationRequest Base64 decoding failed: %v", err)
		return nil, errors.New(104, "Request Error")
	}
	result, err := l.svcCtx.AuthRpc.RegisterInit(l.ctx, &auth.RegisterInitReq{
		Phone:               req.Phone,
		RegistrationRequest: registrationRequest,
	})
	if err != nil {
		l.Errorf("RegisterInit RPC failed: %v", err)
		return nil, err
	}
	return &types.RegisterInitResponse{
		ServerPublicKey:      base64.URLEncoding.EncodeToString(result.ServerPublicKey),
		RegistrationResponse: base64.URLEncoding.EncodeToString(result.RegistrationResponse),
		CredentialIdentifier: base64.URLEncoding.EncodeToString(result.CredentialIdentifier),
	}, nil
}
