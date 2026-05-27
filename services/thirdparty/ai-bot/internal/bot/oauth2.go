package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"ai-bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OAuth2Manager struct {
	tokenURL     string
	clientID     string
	clientSecret string

	mu    sync.RWMutex
	cache *types.TokenCache
}

func NewOAuth2Manager(tokenURL, clientID, clientSecret string) *OAuth2Manager {
	return &OAuth2Manager{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (m *OAuth2Manager) GetAccessToken() (string, error) {
	m.mu.RLock()
	if m.cache != nil && time.Now().Before(m.cache.ExpiresAt.Add(-30*time.Second)) {
		token := m.cache.AccessToken
		m.mu.RUnlock()
		return token, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache != nil && time.Now().Before(m.cache.ExpiresAt.Add(-30*time.Second)) {
		return m.cache.AccessToken, nil
	}

	if m.cache != nil && m.cache.RefreshToken != "" {
		token, err := m.refreshToken(m.cache.RefreshToken)
		if err == nil {
			return token, nil
		}
		logx.Infof("refresh token failed, falling back to client_credentials: %v", err)
	}

	return m.clientCredentials()
}

func (m *OAuth2Manager) refreshToken(refreshToken string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", m.clientID)
	data.Set("client_secret", m.clientSecret)
	data.Set("refresh_token", refreshToken)

	return m.doTokenRequest(data)
}

func (m *OAuth2Manager) clientCredentials() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", m.clientID)
	data.Set("client_secret", m.clientSecret)
	data.Set("scope", "message.read message.write message.reply")

	return m.doTokenRequest(data)
}

func (m *OAuth2Manager) doTokenRequest(data url.Values) (string, error) {
	req, err := http.NewRequest("POST", m.tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var token types.OAuth2Token
	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	// Validate that the token response contains required fields.
	// The gateway may return HTTP 200 with an error body (e.g. {"code":-1,"message":"..."})
	// which would unmarshal to zero values silently.
	if token.AccessToken == "" {
		return "", fmt.Errorf("token response missing access_token, body: %s", string(body))
	}
	if token.ExpiresIn <= 0 {
		return "", fmt.Errorf("token response invalid expires_in=%d, body: %s", token.ExpiresIn, string(body))
	}

	m.cache = &types.TokenCache{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}

	logx.Infof("OAuth2 token obtained, expires at %s", m.cache.ExpiresAt.Format(time.RFC3339))
	return token.AccessToken, nil
}
