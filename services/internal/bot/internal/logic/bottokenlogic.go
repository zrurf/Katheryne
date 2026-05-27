package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotTokenLogic {
	return &BotTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotTokenLogic) BotToken(in *bot.BotTokenReq) (*bot.BotTokenResp, error) {
	l.Infof("BotToken: grant_type=%s, client_id=%s", in.GrantType, in.ClientId)
	switch in.GrantType {
	case "authorization_code":
		return l.handleAuthorizationCode(in)
	case "refresh_token":
		return l.handleRefreshToken(in)
	case "client_credentials":
		return l.handleClientCredentials(in)
	default:
		return nil, fmt.Errorf("unsupported grant_type: %s", in.GrantType)
	}
}

func (l *BotTokenLogic) handleAuthorizationCode(in *bot.BotTokenReq) (*bot.BotTokenResp, error) {
	authData, err := l.svcCtx.OAuthDao.GetAuthCode(l.ctx, in.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}

	if authData["client_id"] != in.ClientId {
		return nil, fmt.Errorf("client_id mismatch")
	}

	l.svcCtx.OAuthDao.ConsumeAuthCode(l.ctx, in.Code)

	botID, clientSecret, err := l.svcCtx.BotDao.GetBotCredentials(l.ctx, in.ClientId)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	if clientSecret != in.ClientSecret {
		return nil, fmt.Errorf("invalid client_secret")
	}

	return l.issueTokens(botID, in.ClientId)
}

func (l *BotTokenLogic) handleRefreshToken(in *bot.BotTokenReq) (*bot.BotTokenResp, error) {
	clientID, _, err := l.svcCtx.OAuthDao.ConsumeRefreshToken(l.ctx, in.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if clientID != in.ClientId {
		return nil, fmt.Errorf("client_id mismatch")
	}

	botID, clientSecret, err := l.svcCtx.BotDao.GetBotCredentials(l.ctx, in.ClientId)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	if clientSecret != in.ClientSecret {
		return nil, fmt.Errorf("invalid client_secret")
	}

	return l.issueTokens(botID, in.ClientId)
}

func (l *BotTokenLogic) handleClientCredentials(in *bot.BotTokenReq) (*bot.BotTokenResp, error) {
	botID, clientSecret, err := l.svcCtx.BotDao.GetBotCredentials(l.ctx, in.ClientId)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	if clientSecret != in.ClientSecret {
		return nil, fmt.Errorf("invalid client_secret")
	}

	return l.issueTokens(botID, in.ClientId)
}

func (l *BotTokenLogic) issueTokens(botID int64, clientID string) (*bot.BotTokenResp, error) {
	accessToken, _ := generateBotToken()
	refreshToken, _ := generateBotToken()
	expiresIn := int64(86400)

	l.svcCtx.OAuthDao.StoreAccessToken(l.ctx, accessToken, botID, clientID, "message.read", expiresIn)
	l.svcCtx.OAuthDao.StoreRefreshToken(l.ctx, refreshToken, clientID, "message.read")

	return &bot.BotTokenResp{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		Scope:        "message.read",
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		BotId:        botID,
	}, nil
}

func generateBotToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}