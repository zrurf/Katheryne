package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

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

	botInfo, err := l.svcCtx.BotDao.GetBotByID(l.ctx, req.BotID, uid)
	if err != nil {
		return nil, err
	}

	return &types.GetBotResp{
		Bot: *botInfo,
	}, nil
}
