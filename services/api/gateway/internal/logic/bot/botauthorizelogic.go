package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotAuthorizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotAuthorizeLogic {
	return &BotAuthorizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotAuthorizeLogic) BotAuthorize(req *types.BotAuthorizeReq) (resp *types.BotAuthorizeResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	result, err := l.svcCtx.BotRpc.Authorize(l.ctx, &botclient.AuthorizeReq{
		ClientId: req.ClientId,
		Scope:    req.Scope,
		ConvId:   convId,
	})
	if err != nil {
		return nil, err
	}

	botInfo := result.Bot
	var botItem *types.BotItem
	if botInfo != nil {
		botItem = &types.BotItem{
			BotId:       strconv.FormatInt(botInfo.BotId, 10),
			Name:        botInfo.Name,
			Description: botInfo.Description,
			Avatar:      botInfo.Avatar,
			Status:      botInfo.Status,
			ClientId:    botInfo.ClientId,
			CreatedAt:   botInfo.CreatedAt,
		}
	}

	return &types.BotAuthorizeResp{
		Bot:            botItem,
		RequestedScope: result.RequestedScope,
		ConvId:         strconv.FormatInt(result.ConvId, 10),
	}, nil
}
