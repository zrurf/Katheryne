package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvLogic {
	return &BotGetConvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetConvLogic) BotGetConv(req *types.BotGetConvReq) (resp *types.BotGetConvResp, err error) {
	return &types.BotGetConvResp{
		ConvID:    req.ConvID,
		Type:      "group",
		Name:      "",
		CreatedAt: 0,
	}, nil
}