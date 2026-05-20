package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotTokenLogic {
	return &BotTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotTokenLogic) BotToken(req *types.BotTokenReq) (resp *types.BotTokenResp, err error) {
	switch req.GrantType {
	case "authorization_code":
		return l.handleAuthorizationCode(req)
	case "refresh_token":
		return l.handleRefreshToken(req)
	case "client_credentials":
		return l.handleClientCredentials(req)
	default:
		return nil, fmt.Errorf("unsupported grant_type: %s", req.GrantType)
	}
}

func (l *BotTokenLogic) handleAuthorizationCode(req *types.BotTokenReq) (*types.BotTokenResp, error) {
	authData, err := l.svcCtx.OAuthDao.GetAuthCode(l.ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}

	if authData["client_id"] != req.ClientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	l.svcCtx.OAuthDao.ConsumeAuthCode(l.ctx, req.Code)

	botID, clientSecret, err := l.svcCtx.BotDao.GetBotCredentials(l.ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	if clientSecret != req.ClientSecret {
		return nil, fmt.Errorf("invalid client_secret")
	}

	return l.issueTokens(botID, req.ClientID)
}

func (l *BotTokenLogic) handleRefreshToken(req *types.BotTokenReq) (*types.BotTokenResp, error) {
	clientID, _, err := l.svcCtx.OAuthDao.ConsumeRefreshToken(l.ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if clientID != req.ClientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	botID, clientSecret, err := l.svcCtx.BotDao.GetBotCredentials(l.ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	if clientSecret != req.ClientSecret {
		return nil, fmt.Errorf("invalid client_secret")
	}

	return l.issueTokens(botID, req.ClientID)
}

func (l *BotTokenLogic) handleClientCredentials(req *types.BotTokenReq) (*types.BotTokenResp, error) {
	botID, clientSecret, err := l.svcCtx.BotDao.GetBotCredentials(l.ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	if clientSecret != req.ClientSecret {
		return nil, fmt.Errorf("invalid client_secret")
	}

	return l.issueTokens(botID, req.ClientID)
}

func (l *BotTokenLogic) issueTokens(botID int64, clientID string) (*types.BotTokenResp, error) {
	accessToken, _ := generateToken()
	refreshToken, _ := generateToken()
	expiresIn := int64(86400)

	l.svcCtx.OAuthDao.StoreAccessToken(l.ctx, accessToken, botID, clientID, "message.read", expiresIn)
	l.svcCtx.OAuthDao.StoreRefreshToken(l.ctx, refreshToken, clientID, "message.read")

	return &types.BotTokenResp{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		Scope:        "message.read",
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		BotID:        botID,
	}, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
