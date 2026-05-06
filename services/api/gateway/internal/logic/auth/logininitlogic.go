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

type LoginInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginInitLogic {
	return &LoginInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginInitLogic) LoginInit(req *types.LoginInitRequest) (resp *types.LoginInitResponse, err error) {
	ke1, err := base64.URLEncoding.DecodeString(req.Ke1)
	if err != nil {
		l.Errorf("KE1 Base64 decoding failed: %v", err)
		return nil, errors.New(104, "Request Error")
	}
	result, err := l.svcCtx.AuthRpc.LoginInit(l.ctx, &auth.LoginInitReq{
		Phone: req.Phone,
		Ke1:   ke1,
	})
	if err != nil {
		l.Errorf("LoginInit RPC failed: %v", err)
		return nil, err
	}
	return &types.LoginInitResponse{
		Ke2:       base64.URLEncoding.EncodeToString(result.Ke2),
		SessionId: result.SessionId,
	}, nil
}
