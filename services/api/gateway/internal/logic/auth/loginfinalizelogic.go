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

type LoginFinalizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginFinalizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginFinalizeLogic {
	return &LoginFinalizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginFinalizeLogic) LoginFinalize(req *types.LoginFinalizeRequest) (resp *types.LoginFinalizeResponse, err error) {
	ke3, err := base64.URLEncoding.DecodeString(req.Ke3)
	if err != nil {
		l.Errorf("KE3 Base64 decoding failed: %v", err)
		return nil, errors.New(104, "Request Error")
	}
	result, err := l.svcCtx.AuthRpc.LoginFinalize(l.ctx, &auth.LoginFinalizeReq{
		Ke3:       ke3,
		SessionId: req.SessionId,
	})
	if err != nil {
		l.Errorf("LoginFinalize RPC failed: %v", err)
		return nil, err
	}
	return &types.LoginFinalizeResponse{
		UID:          result.Uid,
		Need2FA:      result.Need_2Fa,
		TFAToken:     result.TfaToken,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
	}, nil
}
