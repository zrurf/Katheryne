// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oauth2

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
