package oauth2

import (
	"context"
	"fmt"
	"strings"

	"bot/internal/svc"
	"bot/internal/types"

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

func (l *AuthorizeLogic) Authorize(req *types.AuthorizeReq) (resp *types.AuthorizeResp, err error) {
	bot, err := l.svcCtx.BotDao.GetBotByClientID(l.ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	scopes := strings.Split(strings.TrimSpace(req.Scope), ",")
	if req.Scope == "" {
		scopes = []string{"message.read"}
	}

	return &types.AuthorizeResp{
		Bot:            *bot,
		RequestedScope: scopes,
		ConvID:         req.ConvID,
	}, nil
}