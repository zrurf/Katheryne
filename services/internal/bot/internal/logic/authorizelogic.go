package logic

import (
	"context"
	"fmt"
	"strings"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthorizeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthorizeLogic {
	return &AuthorizeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthorizeLogic) Authorize(in *bot.AuthorizeReq) (*bot.AuthorizeResp, error) {
	botInfo, err := l.svcCtx.BotDao.GetBotByClientID(l.ctx, in.ClientId)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	scopes := strings.Split(strings.TrimSpace(in.Scope), ",")
	if in.Scope == "" {
		scopes = []string{"message.read"}
	}

	return &bot.AuthorizeResp{
		Bot:            botInfo,
		RequestedScope: scopes,
		ConvId:         in.ConvId,
	}, nil
}