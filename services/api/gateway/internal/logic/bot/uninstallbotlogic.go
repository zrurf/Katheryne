package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UninstallBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUninstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UninstallBotLogic {
	return &UninstallBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UninstallBotLogic) UninstallBot(req *types.UninstallBotReq) (resp *types.UninstallBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	_, err = l.svcCtx.BotRpc.UninstallBot(l.ctx, &botclient.UninstallBotReq{
		BotId:  botId,
		ConvId: convId,
		Uid:    uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.UninstallBotResp{}, nil
}
