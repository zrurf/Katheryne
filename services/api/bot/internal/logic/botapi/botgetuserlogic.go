package botapi

import (
	"context"
	"fmt"
	"strings"

	"bot/internal/middleware"
	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetUserLogic {
	return &BotGetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetUserLogic) BotGetUser(req *types.BotGetUserReq) (resp *types.BotGetUserResp, err error) {
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	scopeHasUser := strings.Contains(auth.Scope, "user.read") || strings.Contains(auth.Scope, "*")
	if !scopeHasUser {
		return nil, fmt.Errorf("insufficient scope: user.read required")
	}

	name, avatar, err := l.svcCtx.InstallationDao.GetUserInfo(l.ctx, req.UID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &types.BotGetUserResp{
		UID:    req.UID,
		Name:   name,
		Avatar: avatar,
	}, nil
}