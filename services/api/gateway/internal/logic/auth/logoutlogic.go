// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"auth/auth"
	"context"
	"strings"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutRequest) (resp *types.EmptyReponse, err error) {
	parts := strings.Fields(req.Authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		l.Errorf("Incorrect Authorization: %v", req.Authorization)
		return nil, errors.New(100, "Unauthorized")
	}
	token := parts[1]
	_, err = l.svcCtx.AuthRpc.Logout(l.ctx, &auth.LogoutReq{
		AccessToken: token,
	})
	if err != nil {
		l.Errorf("Logout RPC failed: %v", err)
		return nil, err
	}
	return &types.EmptyReponse{}, err
}
