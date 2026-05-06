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

type LoginFinalizeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginFinalizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginFinalizeLogic {
	return &LoginFinalizeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginFinalizeLogic) LoginFinalize(in *auth.LoginFinalizeReq) (*auth.LoginFinalizeResp, error) {

	ke3, err := l.svcCtx.OpaqueSvc.GetServer().Deserialize.KE3(in.Ke3)
	if err != nil {
		l.Logger.Infof("Failed to deserialize KE3: %v", err)
		return nil, errors.New(104, "Request Error")
	}

	sessionData, err := l.svcCtx.SessionDao.GetLoginSession(l.ctx, in.SessionId)
	if err != nil {
		l.Logger.Infof("Failed to get login session: %v", err)
		return nil, errors.New(301, "Login session not found")
	}

	// 调用 LoginFinish 校验 MAC
	if err := l.svcCtx.OpaqueSvc.GetServer().LoginFinish(ke3, sessionData.MAC); err != nil {
		l.Logger.Infof("Invalid MAC: %v", err)
		return nil, errors.New(11, "Login failed")
	}

	// 认证通过，获取用户 ID
	uid, _, err := l.svcCtx.UserDao.GetUserRecord(l.ctx, sessionData.User)
	if err != nil {
		l.Logger.Errorf("Failed to get user: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}

	// 生成会话 token 并写入内存数据库
	accessToken, refreshToken, err := utils.GenerateAndSaveToken(l.ctx, l.svcCtx.OpaqueSvc, l.svcCtx.SessionDao, uid, l.svcCtx.Config.AccessTokenExpireSeconds, l.svcCtx.Config.RefreshTokenExpireSeconds, l.svcCtx.Config.MaxTokenRetries, l.svcCtx.Config.SessionTokenLength)
	if err != nil {
		l.Logger.Errorf("Failed to generate session token: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}

	// 更新最后登录时间
	_ = l.svcCtx.UserDao.UpdateLastLogin(l.ctx, uid)

	return &auth.LoginFinalizeResp{
		Uid:          uid,
		Need_2Fa:     false,
		TfaToken:     "",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(l.svcCtx.Config.AccessTokenExpireSeconds) * time.Second).UnixMilli(),
	}, nil
}
