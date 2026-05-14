package developer

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

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

	l.svcCtx.Redis.HDel(l.ctx, "bots", fmt.Sprintf("%d", req.BotID))
	l.svcCtx.Redis.SRem(l.ctx, fmt.Sprintf("user_bots:%d", uid), req.BotID)

	return &types.DeleteBotResp{}, nil
}