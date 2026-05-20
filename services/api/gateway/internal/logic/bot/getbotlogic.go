package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotLogic {
	return &GetBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotLogic) GetBot(req *types.GetBotReq) (resp *types.GetBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	result, err := l.svcCtx.BotRpc.GetBot(l.ctx, &botclient.GetBotReq{
		BotId: botId,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}
	info := result.Bot
	if info == nil {
		return &types.GetBotResp{}, nil
	}
	return &types.GetBotResp{
		BotId:       strconv.FormatInt(info.BotId, 10),
		Name:        info.Name,
		Description: info.Description,
		Avatar:      info.Avatar,
		Status:      info.Status,
		ClientId:    info.ClientId,
		WebhookUrl:  info.WebhookUrl,
		CreatedAt:   info.CreatedAt,
	}, nil
}
