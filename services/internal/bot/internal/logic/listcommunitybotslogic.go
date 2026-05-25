package logic

import (
	"context"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommunityBotsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCommunityBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommunityBotsLogic {
	return &ListCommunityBotsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListCommunityBotsLogic) ListCommunityBots(in *bot.ListCommunityBotsReq) (*bot.ListCommunityBotsResp, error) {
	section := in.Section
	if section == "" {
		section = "all"
	}

	var hostedBots []*bot.CommunityBotItem
	var templates []*bot.CommunityBotItem

	// Fetch hosted instances (non-self-hosted bots)
	if section == "all" || section == "hosted" {
		instances, err := l.svcCtx.InstanceDao.ListHostedInstances(l.ctx, in.Keyword)
		if err != nil {
			l.Logger.Errorf("ListHostedInstances err: %v", err)
		} else {
			for _, inst := range instances {
				desc := ""
				if inst.Template != nil {
					desc = inst.Template.Description
				}
				hostedBots = append(hostedBots, &bot.CommunityBotItem{
					Type:           "hosted",
					InstanceId:     inst.InstanceId,
					BotId:          inst.BotId,
					Name:           inst.Name,
					Avatar:         inst.Avatar,
					Description:    desc,
					HostedBy:       inst.HostedBy,
					Status:         inst.Status,
					InstalledCount: 0,
				})
			}
		}
	}

	// Fetch templates
	if section == "all" || section == "templates" {
		tmpls, err := l.svcCtx.TemplateDao.ListCommunityTemplates(l.ctx, in.Keyword, in.Category)
		if err != nil {
			l.Logger.Errorf("ListCommunityTemplates err: %v", err)
		} else {
			for _, tmpl := range tmpls {
				templates = append(templates, &bot.CommunityBotItem{
					Type:        "template",
					TemplateId:  tmpl.TemplateId,
					Name:        tmpl.Name,
					Avatar:      tmpl.Avatar,
					Description: tmpl.Description,
					Category:    tmpl.Category,
					Tags:        tmpl.Tags,
					IsOfficial:  tmpl.IsOfficial,
					Status:      tmpl.Status,
				})
			}
		}
	}

	if hostedBots == nil {
		hostedBots = []*bot.CommunityBotItem{}
	}
	if templates == nil {
		templates = []*bot.CommunityBotItem{}
	}

	return &bot.ListCommunityBotsResp{
		HostedBots: hostedBots,
		Templates:  templates,
	}, nil
}
