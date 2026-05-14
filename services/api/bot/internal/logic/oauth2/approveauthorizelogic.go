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

type ApproveAuthorizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveAuthorizeLogic {
	return &ApproveAuthorizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveAuthorizeLogic) ApproveAuthorize(req *types.ApproveAuthorizeReq) (resp *types.ApproveAuthorizeResp, err error) {
	codeBytes := make([]byte, 32)
	rand.Read(codeBytes)
	code := hex.EncodeToString(codeBytes)

	authData := map[string]interface{}{
		"client_id":    req.ClientID,
		"redirect_uri": req.RedirectURI,
		"scope":        req.Scope,
		"conv_id":      req.ConvID,
	}
	data, _ := json.Marshal(authData)
	l.svcCtx.Redis.Set(l.ctx, "oauth2:code:"+code, data, 10*time.Minute)

	redirectURL := fmt.Sprintf("%s?code=%s", req.RedirectURI, code)
	if req.State != "" {
		redirectURL += "&state=" + req.State
	}

	return &types.ApproveAuthorizeResp{
		RedirectURL: redirectURL,
	}, nil
}