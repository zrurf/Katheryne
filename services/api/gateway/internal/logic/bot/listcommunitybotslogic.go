package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommunityBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCommunityBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommunityBotsLogic {
	return &ListCommunityBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCommunityBotsLogic) ListCommunityBots(req *types.ListCommunityBotsReq) (resp *types.ListCommunityBotsResp, err error) {
	result, err := l.svcCtx.BotRpc.ListCommunityBots(l.ctx, &botclient.ListCommunityBotsReq{
		Keyword: req.Keyword,
	})
	if err != nil {
		return nil, err
	}

	hostedBots := convertCommunityBots(result.HostedBots)
	templates := convertCommunityBots(result.Templates)

	return &types.ListCommunityBotsResp{
		HostedBots: hostedBots,
		Templates:  templates,
	}, nil
}

func convertCommunityBots(items []*botclient.CommunityBotItem) []types.CommunityBotItem {
	list := make([]types.CommunityBotItem, 0, len(items))
	for _, item := range items {
		b := types.CommunityBotItem{
			Type:           item.Type,
			Name:           item.Name,
			Avatar:         item.Avatar,
			Description:    item.Description,
			IsOfficial:     item.IsOfficial,
			Status:         item.Status,
			InstalledCount: item.InstalledCount,
		}
		if item.InstanceId > 0 {
			b.InstanceId = strconv.FormatInt(item.InstanceId, 10)
		}
		if item.BotId > 0 {
			b.BotId = strconv.FormatInt(item.BotId, 10)
		}
		if item.TemplateId > 0 {
			b.TemplateId = strconv.FormatInt(item.TemplateId, 10)
		}
		if item.HostedBy > 0 {
			b.HostedBy = strconv.FormatInt(item.HostedBy, 10)
		}
		if item.Category != "" {
			b.Category = item.Category
		}
		if item.Tags != nil {
			b.Tags = item.Tags
		}
		list = append(list, b)
	}
	return list
}