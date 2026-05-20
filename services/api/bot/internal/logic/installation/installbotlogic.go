package installation

import (
	"context"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InstallBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InstallBotLogic {
	return &InstallBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InstallBotLogic) InstallBot(req *types.InstallBotReq) (resp *types.InstallBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	convInfo, err := l.svcCtx.InstallationDao.GetConversation(l.ctx, req.ConvID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if convInfo.ConvType == "GROUP" && convInfo.GroupID > 0 {
		if err := l.svcCtx.InstallationDao.CheckGroupMemberRole(l.ctx, convInfo.GroupID, uid, "OWNER", "ADMIN"); err != nil {
			return nil, fmt.Errorf("only group owner or admin can install bots")
		}
	}

	if err := l.svcCtx.InstallationDao.Install(l.ctx, req.BotID, req.ConvID, convInfo.ConvType, req.Permissions, uid); err != nil {
		return nil, fmt.Errorf("failed to install bot: %v", err)
	}

	l.svcCtx.OAuthDao.AddConvBotCache(l.ctx, req.ConvID, req.BotID)

	return &types.InstallBotResp{}, nil
}