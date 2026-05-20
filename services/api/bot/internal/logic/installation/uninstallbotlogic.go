package installation

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

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

	convInfo, err := l.svcCtx.InstallationDao.GetConversation(l.ctx, req.ConvID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if convInfo.ConvType == "GROUP" && convInfo.GroupID > 0 {
		if err := l.svcCtx.InstallationDao.CheckGroupMemberRole(l.ctx, convInfo.GroupID, uid, "OWNER", "ADMIN"); err != nil {
			return nil, fmt.Errorf("only group owner or admin can uninstall bots")
		}
	}

	if err := l.svcCtx.InstallationDao.Uninstall(l.ctx, req.BotID, req.ConvID); err != nil {
		return nil, fmt.Errorf("failed to uninstall bot: %v", err)
	}

	l.svcCtx.OAuthDao.RemoveConvBotCache(l.ctx, req.ConvID, req.BotID)

	return &types.UninstallBotResp{}, nil
}