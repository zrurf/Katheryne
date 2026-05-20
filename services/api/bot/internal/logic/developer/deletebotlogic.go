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

	rowsAffected, err := l.svcCtx.BotDao.DeleteBot(l.ctx, req.BotID, uid)
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	return &types.DeleteBotResp{}, nil
}