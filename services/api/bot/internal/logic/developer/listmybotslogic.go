package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyBotsLogic {
	return &ListMyBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyBotsLogic) ListMyBots() (resp *types.ListMyBotsResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	list, err := l.svcCtx.BotDao.ListBotsByOwner(l.ctx, uid)
	if err != nil {
		return nil, err
	}

	return &types.ListMyBotsResp{List: list}, nil
}