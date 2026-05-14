package oauth2

import (
	"context"
	"encoding/json"
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
	var bot types.BotInfo
	botData, err := l.svcCtx.Redis.HGetAll(l.ctx, "bots").Result()
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	for _, data := range botData {
		var b types.BotInfo
		json.Unmarshal([]byte(data), &b)
		if b.ClientID == req.ClientID {
			bot = b
			break
		}
	}

	if bot.BotID == 0 {
		return nil, fmt.Errorf("bot not found")
	}

	scopes := strings.Split(req.Scope, ",")

	return &types.AuthorizeResp{
		Bot:            bot,
		RequestedScope: scopes,
		ConvID:         req.ConvID,
	}, nil
}