package developer

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotInstallationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotInstallationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotInstallationsLogic {
	return &GetBotInstallationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotInstallationsLogic) GetBotInstallations(req *types.GetBotInstallationsReq) (resp *types.GetBotInstallationsResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, req.BotID, uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	list, err := l.svcCtx.InstallationDao.ListBotInstallations(l.ctx, req.BotID)
	if err != nil {
		return nil, err
	}

	return &types.GetBotInstallationsResp{List: list}, nil
}
