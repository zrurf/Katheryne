package oauth2

import (
	"context"
	"fmt"

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
	uid := l.ctx.Value("uid").(int64)

	code, err := l.svcCtx.OAuthDao.StoreAuthCode(l.ctx, req.ClientID, req.RedirectURI, req.Scope, uid, req.ConvID)
	if err != nil {
		return nil, fmt.Errorf("failed to store auth code: %v", err)
	}

	redirectURL := fmt.Sprintf("%s?code=%s", req.RedirectURI, code)
	if req.State != "" {
		redirectURL += "&state=" + req.State
	}

	return &types.ApproveAuthorizeResp{
		RedirectURL: redirectURL,
	}, nil
}