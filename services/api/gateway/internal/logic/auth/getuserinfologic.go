// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"
	"strings"
	"user/userclient"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.UserInfoRequest) (resp *types.UserInfoResponse, err error) {
	var token string
	parts := strings.Fields(req.Authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		token = ""
	} else {
		token = parts[1]
	}

	r, err := l.svcCtx.UserRpc.GetUserByUID(l.ctx, &userclient.GetUserByUIDReq{
		Uid:   req.Uid,
		Token: token,
	})

	if err != nil {
		l.Errorf("GetUserByUID RPC error: %v", err)
		return nil, err
	}

	return &types.UserInfoResponse{
		UID:       r.User.Uid,
		Name:      r.User.Name,
		Avatar:    r.User.Avatar,
		Status:    r.User.Status,
		CreatedAt: r.User.CreatedAt,
		UpdatedAt: r.User.UpdatedAt,
		LastLogin: r.User.LastLogin,
	}, nil
}
