package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TokenLogic {
	return &TokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TokenLogic) Token(req *types.TokenRequest) (resp *types.TokenResponse, err error) {
	switch req.GrantType {
	case "authorization_code":
		return l.handleAuthorizationCode(req)
	case "refresh_token":
		return l.handleRefreshToken(req)
	default:
		return nil, fmt.Errorf("unsupported grant_type: %s", req.GrantType)
	}
}

func (l *TokenLogic) handleAuthorizationCode(req *types.TokenRequest) (*types.TokenResponse, error) {
	key := fmt.Sprintf("oauth2:code:%s", req.Code)
	clientID, err := l.svcCtx.Redis.Get(l.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}
	if clientID != req.ClientId {
		return nil, fmt.Errorf("client_id mismatch")
	}
	l.svcCtx.Redis.Del(l.ctx, key)

	accessToken := generateToken()
	refreshToken := generateToken()
	expiresIn := int64(7200)

	tokenKey := fmt.Sprintf("oauth2:token:%s", accessToken)
	tokenData := map[string]interface{}{
		"client_id":     req.ClientId,
		"scope":         req.Scope,
		"refresh_token": refreshToken,
		"expires_at":    time.Now().Unix() + expiresIn,
	}
	l.svcCtx.Redis.HMSet(l.ctx, tokenKey, tokenData)
	l.svcCtx.Redis.Expire(l.ctx, tokenKey, time.Duration(expiresIn)*time.Second)

	return &types.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		Scope:        req.Scope,
	}, nil
}

func (l *TokenLogic) handleRefreshToken(req *types.TokenRequest) (*types.TokenResponse, error) {
	accessToken := generateToken()
	newRefreshToken := generateToken()
	expiresIn := int64(7200)

	tokenKey := fmt.Sprintf("oauth2:token:%s", accessToken)
	tokenData := map[string]interface{}{
		"client_id":     req.ClientId,
		"scope":         req.Scope,
		"refresh_token": newRefreshToken,
		"expires_at":    time.Now().Unix() + expiresIn,
	}
	l.svcCtx.Redis.HMSet(l.ctx, tokenKey, tokenData)
	l.svcCtx.Redis.Expire(l.ctx, tokenKey, time.Duration(expiresIn)*time.Second)

	return &types.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		Scope:        req.Scope,
	}, nil
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
