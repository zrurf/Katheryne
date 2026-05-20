package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveAuthorizeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApproveAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveAuthorizeLogic {
	return &ApproveAuthorizeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApproveAuthorizeLogic) ApproveAuthorize(in *bot.ApproveAuthorizeReq) (*bot.ApproveAuthorizeResp, error) {
	code, err := l.svcCtx.OAuthDao.StoreAuthCode(l.ctx, in.ClientId, in.RedirectUri, in.Scope, in.Uid, in.ConvId)
	if err != nil {
		return nil, fmt.Errorf("failed to store auth code: %v", err)
	}

	redirectURL := fmt.Sprintf("%s?code=%s", in.RedirectUri, code)
	if in.State != "" {
		redirectURL += "&state=" + in.State
	}

	return &bot.ApproveAuthorizeResp{
		RedirectUrl: redirectURL,
	}, nil
}