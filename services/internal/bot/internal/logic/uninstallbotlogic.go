package logic

import (
	"context"
	"fmt"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UninstallBotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUninstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UninstallBotLogic {
	return &UninstallBotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UninstallBotLogic) UninstallBot(in *bot.UninstallBotReq) (*bot.UninstallBotResp, error) {
	convType, groupID, err := l.svcCtx.InstDao.GetConversation(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if convType == "GROUP" && groupID > 0 {
		if err := l.svcCtx.InstDao.CheckGroupMemberRole(l.ctx, groupID, in.Uid, "OWNER", "ADMIN"); err != nil {
			return nil, fmt.Errorf("only group owner or admin can uninstall bots")
		}
	}

	if err := l.svcCtx.InstDao.Uninstall(l.ctx, in.BotId, in.ConvId); err != nil {
		return nil, fmt.Errorf("failed to uninstall bot: %v", err)
	}

	l.svcCtx.OAuthDao.RemoveConvBotCache(l.ctx, in.ConvId, in.BotId)

	return &bot.UninstallBotResp{}, nil
}