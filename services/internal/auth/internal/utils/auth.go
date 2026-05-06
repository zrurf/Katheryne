package utils

import (
	"auth/internal/dao"
	"auth/internal/module"
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

// 生成并保存Toekn对
func GenerateAndSaveToken(ctx context.Context, opaque *module.OpaqueService, sessionDao *dao.SessionDao, uid int64, accessTokenExpireSeconds int, refreshTokenExpireSeconds int, maxTokenRetries int, sessionTokenLength int) (string, string, error) {
	var accessToken, refreshToken string
	var accessOK = false
	var refreshOK = false
	for range maxTokenRetries {
		accessToken = opaque.GenerateToken(sessionTokenLength)
		if res, err := sessionDao.HasAccessToken(ctx, accessToken); err == nil && !res {
			accessOK = true
			break
		}
	}

	for range maxTokenRetries {
		refreshToken = opaque.GenerateToken(sessionTokenLength)
		if res, err := sessionDao.HasRefreshToken(ctx, refreshToken); err == nil && !res {
			refreshOK = true
			break
		}
	}

	if accessOK && refreshOK {
		err := sessionDao.SaveAccessToken(ctx, accessToken, uid, accessTokenExpireSeconds)
		if err != nil {
			log.Err(err).Msg("failed to save access token")
			return "", "", err
		}
		err = sessionDao.SaveRefreshToken(ctx, refreshToken, uid, refreshTokenExpireSeconds)
		if err != nil {
			sessionDao.DelAccessToken(ctx, accessToken)
			log.Err(err).Msg("failed to save refresh token")
			return "", "", err
		}
		err = sessionDao.SaveTokenMap(ctx, accessToken, refreshToken, accessTokenExpireSeconds)
		if err != nil {
			sessionDao.DelAccessToken(ctx, accessToken)
			sessionDao.DelRefreshToken(ctx, refreshToken)
			log.Err(err).Msg("failed to save token map")
			return "", "", err
		}
		return accessToken, refreshToken, nil
	} else {
		log.Warn().Msg("failed to generate unique tokens")
		return "", "", fmt.Errorf("failed to generate unique tokens")
	}
}
