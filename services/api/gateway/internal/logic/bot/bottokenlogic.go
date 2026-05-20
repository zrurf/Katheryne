package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

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
	result, err := l.svcCtx.BotRpc.BotToken(l.ctx, &botclient.BotTokenReq{
		GrantType:    req.GrantType,
		ClientId:     req.ClientId,
		ClientSecret: req.ClientSecret,
		Code:         req.Code,
		RedirectUri:  req.RedirectUri,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &types.BotTokenResp{
		AccessToken:  result.AccessToken,
		TokenType:    result.TokenType,
		Scope:        result.Scope,
		ExpiresIn:    result.ExpiresIn,
		RefreshToken: result.RefreshToken,
		BotId:        strconv.FormatInt(result.BotId, 10),
	}, nil
}
