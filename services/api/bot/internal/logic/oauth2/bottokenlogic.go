package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

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
	default:
		return nil, fmt.Errorf("unsupported grant_type: %s", req.GrantType)
	}
}

func (l *BotTokenLogic) handleAuthorizationCode(req *types.BotTokenReq) (*types.BotTokenResp, error) {
	data, err := l.svcCtx.Redis.Get(l.ctx, "oauth2:code:"+req.Code).Result()
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}

	var authData map[string]interface{}
	json.Unmarshal([]byte(data), &authData)

	if authData["client_id"] != req.ClientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	l.svcCtx.Redis.Del(l.ctx, "oauth2:code:"+req.Code)

	accessToken, _ := generateToken()
	refreshToken, _ := generateToken()

	expiresIn := int64(3600)
	l.svcCtx.Redis.Set(l.ctx, "oauth2:access_token:"+accessToken, req.ClientID, time.Duration(expiresIn)*time.Second)
	l.svcCtx.Redis.Set(l.ctx, "oauth2:refresh_token:"+refreshToken, req.ClientID, 30*24*time.Hour)

	var botID int64
	botData, _ := l.svcCtx.Redis.HGetAll(l.ctx, "bots").Result()
	for _, bd := range botData {
		var b types.BotInfo
		json.Unmarshal([]byte(bd), &b)
		if b.ClientID == req.ClientID {
			botID = b.BotID
			break
		}
	}

	return &types.BotTokenResp{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		BotID:        botID,
	}, nil
}

func (l *BotTokenLogic) handleRefreshToken(req *types.BotTokenReq) (*types.BotTokenResp, error) {
	clientID, err := l.svcCtx.Redis.Get(l.ctx, "oauth2:refresh_token:"+req.RefreshToken).Result()
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	accessToken, _ := generateToken()
	refreshToken, _ := generateToken()

	expiresIn := int64(3600)
	l.svcCtx.Redis.Del(l.ctx, "oauth2:refresh_token:"+req.RefreshToken)
	l.svcCtx.Redis.Set(l.ctx, "oauth2:access_token:"+accessToken, clientID, time.Duration(expiresIn)*time.Second)
	l.svcCtx.Redis.Set(l.ctx, "oauth2:refresh_token:"+refreshToken, clientID, 30*24*time.Hour)

	var botID int64
	botData, _ := l.svcCtx.Redis.HGetAll(l.ctx, "bots").Result()
	for _, bd := range botData {
		var b types.BotInfo
		json.Unmarshal([]byte(bd), &b)
		if b.ClientID == clientID {
			botID = b.BotID
			break
		}
	}

	return &types.BotTokenResp{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
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
