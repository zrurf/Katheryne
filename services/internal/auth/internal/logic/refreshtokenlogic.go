package logic

import (
	"context"
	"time"

	"auth/auth"
	"auth/internal/svc"
	"auth/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshTokenLogic) RefreshToken(in *auth.RefreshTokenReq) (*auth.RefreshTokenResp, error) {
	if len(in.RefreshToken) == 0 {
		return nil, errors.New(104, "Request Error")
	}
	exist, err := l.svcCtx.SessionDao.HasRefreshToken(l.ctx, in.RefreshToken)
	if err != nil {
		l.Errorf("Session DAO Error: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}
	if !exist {
		return nil, errors.New(101, "Token does not exist")
	}
	uid, err := l.svcCtx.SessionDao.GetUidByRefreshToken(l.ctx, in.RefreshToken)
	if err != nil {
		l.Errorf("Session DAO Error: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}
	err = l.svcCtx.SessionDao.DelRefreshToken(l.ctx, in.RefreshToken)
	if err != nil {
		l.Errorf("Session DAO Error: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}
	accessToken, err := l.svcCtx.SessionDao.GetTokenMap(l.ctx, in.RefreshToken)
	if err != nil {
		l.Errorf("Session DAO Error: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}
	err = l.svcCtx.SessionDao.DelAccessToken(l.ctx, accessToken)
	if err != nil {
		l.Errorf("Session DAO Error: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}
	newAccessTk, newRefreshTk, err := utils.GenerateAndSaveToken(l.ctx, l.svcCtx.OpaqueSvc, l.svcCtx.SessionDao, uid, l.svcCtx.Config.AccessTokenExpireSeconds, l.svcCtx.Config.RefreshTokenExpireSeconds, l.svcCtx.Config.MaxTokenRetries, l.svcCtx.Config.SessionTokenLength)
	if err != nil {
		l.Errorf("Failed to generate session token: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}
	return &auth.RefreshTokenResp{
		AccessToken:  newAccessTk,
		RefreshToken: newRefreshTk,
		ExpiresAt:    time.Now().Add(time.Duration(l.svcCtx.Config.AccessTokenExpireSeconds) * time.Second).UnixMilli(),
	}, nil
}
