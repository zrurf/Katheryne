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

type AuthorizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthorizeLogic {
	return &AuthorizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthorizeLogic) Authorize(req *types.AuthorizeRequest) (resp *types.EmptyReponse, err error) {
	if req.ResponseType != "code" {
		return nil, fmt.Errorf("unsupported response_type: %s", req.ResponseType)
	}

	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return nil, err
	}
	code := hex.EncodeToString(codeBytes)

	key := fmt.Sprintf("oauth2:code:%s", code)
	err = l.svcCtx.Redis.Set(l.ctx, key, req.ClientId, 10*time.Minute).Err()
	if err != nil {
		l.Errorf("Failed to store auth code: %v", err)
		return nil, err
	}

	return &types.EmptyReponse{}, nil
}
