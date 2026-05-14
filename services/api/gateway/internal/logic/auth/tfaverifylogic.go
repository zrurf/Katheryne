// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"auth/authclient"
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TFAVerifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTFAVerifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TFAVerifyLogic {
	return &TFAVerifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TFAVerifyLogic) TFAVerify(req *types.TFAVerifyRequest) (resp *types.TFAVerifyResponse, err error) {
	result, err := l.svcCtx.AuthRpc.TFAVerify(l.ctx, &authclient.TFAVerifyReq{
		TfaToken: req.TFAToken,
		Code:     req.Code,
	})
	if err != nil {
		l.Errorf("TFAVerify RPC failed: %v", err)
		return nil, err
	}
	return &types.TFAVerifyResponse{
		UID:          strconv.FormatInt(result.Uid, 10),
		AccessToken:  result.AccessToken,
		ExpiresAt:    result.ExpiresAt,
		RefreshToken: result.RefreshToken,
	}, nil
}
