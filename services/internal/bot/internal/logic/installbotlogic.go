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

	botID := in.BotId
	var newInstanceID int64

	// If template_id is provided, this is a self-hosted installation:
	// create a bot instance first, then install
	if in.TemplateId > 0 {
		// Create a new bot instance from template
		createReq := &bot.CreateBotInstanceReq{
			Uid:            in.Uid,
			TemplateId:     in.TemplateId,
			Name:           "", // will be filled from template
			Avatar:         "",
			IsSelfHosted:   true,
			ModelProvider:  in.ModelProvider,
			ModelName:      in.ModelName,
			ApiKey:         in.ApiKey,
			KbConfig:       in.KbConfig,
		}

		// Load template to get name/avatar
		tmpl, err := l.svcCtx.TemplateDao.GetTemplateByID(l.ctx, in.TemplateId)
		if err == nil {
			if createReq.Name == "" {
				createReq.Name = tmpl.Name
			}
			if createReq.Avatar == "" {
				createReq.Avatar = tmpl.Avatar
			}
		}

		resp, err := l.svcCtx.InstanceDao.CreateInstance(l.ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create bot instance: %v", err)
		}
		botID = resp.BotId
		newInstanceID = resp.InstanceId
	} else {
		// Installing a hosted/official bot — validate it exists and is ACTIVE
		exists, err := l.svcCtx.BotDao.BotExists(l.ctx, botID)
		if err != nil {
			l.Errorf("check bot existence failed: bot_id=%d, err=%v", botID, err)
			return nil, fmt.Errorf("failed to verify bot: %v", err)
		}
		if !exists {
			l.Errorf("bot not found or inactive: bot_id=%d", botID)
			return nil, fmt.Errorf("bot not found or inactive: bot_id=%d", botID)
		}
	}

	if err := l.svcCtx.InstDao.Install(l.ctx, botID, in.ConvId, convType, in.Permissions, in.Uid); err != nil {
		return nil, fmt.Errorf("failed to install bot: %v", err)
	}

	l.svcCtx.OAuthDao.AddConvBotCache(l.ctx, in.ConvId, botID)

	return &bot.InstallBotResp{InstanceId: newInstanceID}, nil
}