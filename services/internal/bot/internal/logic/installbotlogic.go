package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type InstallBotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InstallBotLogic {
	return &InstallBotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InstallBotLogic) InstallBot(in *bot.InstallBotReq) (*bot.InstallBotResp, error) {
	convType, groupID, err := l.svcCtx.InstDao.GetConversation(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if convType == "GROUP" && groupID > 0 {
		if err := l.svcCtx.InstDao.CheckGroupMemberRole(l.ctx, groupID, in.Uid, "OWNER", "ADMIN"); err != nil {
			return nil, fmt.Errorf("only group owner or admin can install bots")
		}
	}

	if err := l.svcCtx.InstDao.Install(l.ctx, in.BotId, in.ConvId, convType, in.Permissions, in.Uid); err != nil {
		return nil, fmt.Errorf("failed to install bot: %v", err)
	}

	l.svcCtx.OAuthDao.AddConvBotCache(l.ctx, in.ConvId, in.BotId)

	return &bot.InstallBotResp{}, nil
}