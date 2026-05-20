package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotLogic {
	return &CreateBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateBotLogic) CreateBot(req *types.CreateBotReq) (resp *types.CreateBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	return l.svcCtx.BotDao.CreateBot(l.ctx, uid, req)
}
