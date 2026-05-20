package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBotLogic {
	return &DeleteBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBotLogic) DeleteBot(req *types.DeleteBotReq) (resp *types.DeleteBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	_, err = l.svcCtx.BotRpc.DeleteBot(l.ctx, &botclient.DeleteBotReq{
		BotId: botId,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.DeleteBotResp{}, nil
}
