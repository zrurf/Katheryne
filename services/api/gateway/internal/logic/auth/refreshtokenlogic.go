// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"auth/auth"
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	result, err := l.svcCtx.AuthRpc.RefreshToken(l.ctx, &auth.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		l.Errorf("RefreshToken RPC failed: %v", err)
		return nil, err
	}
	return &types.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		ExpiresAt:    result.ExpiresAt,
		RefreshToken: result.RefreshToken,
	}, nil
}
