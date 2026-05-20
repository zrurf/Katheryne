package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotLogic {
	return &UpdateBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBotLogic) UpdateBot(req *types.UpdateBotReq) (resp *types.UpdateBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	_, err = l.svcCtx.BotRpc.UpdateBot(l.ctx, &botclient.UpdateBotReq{
		Uid:         uid,
		BotId:       botId,
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		WebhookUrl:  req.WebhookUrl,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateBotResp{}, nil
}
