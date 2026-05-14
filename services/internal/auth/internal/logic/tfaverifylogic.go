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

type TFAVerifyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTFAVerifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TFAVerifyLogic {
	return &TFAVerifyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TFAVerifyLogic) TFAVerify(in *auth.TFAVerifyReq) (*auth.TFAVerifyResp, error) {
	if in.TfaToken == "" || in.Code == "" {
		return nil, errors.New(104, "Request Error")
	}

	// 从 Redis 获取 2FA 对应的 uid
	uid, err := l.svcCtx.SessionDao.GetUidBy2FAToken(l.ctx, in.TfaToken)
	if err != nil {
		l.Logger.Infof("Failed to get 2FA session: %v", err)
		return nil, errors.New(301, "2FA session not found")
	}

	// TODO: 实际项目中这里需要调用 TOTP 验证逻辑
	// 目前简化处理，直接通过

	// 生成会话 token
	accessToken, refreshToken, err := utils.GenerateAndSaveToken(
		l.ctx,
		l.svcCtx.OpaqueSvc,
		l.svcCtx.SessionDao,
		uid,
		l.svcCtx.Config.AccessTokenExpireSeconds,
		l.svcCtx.Config.RefreshTokenExpireSeconds,
		l.svcCtx.Config.MaxTokenRetries,
		l.svcCtx.Config.SessionTokenLength,
	)
	if err != nil {
		l.Logger.Errorf("Failed to generate session token: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}

	// 清理 2FA 会话
	_ = l.svcCtx.SessionDao.Del2FAToken(l.ctx, in.TfaToken)

	return &auth.TFAVerifyResp{
		Uid:          uid,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(l.svcCtx.Config.AccessTokenExpireSeconds) * time.Second).UnixMilli(),
	}, nil
}
