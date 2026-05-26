package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyTemplatesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyTemplatesLogic {
	return &ListMyTemplatesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyTemplatesLogic) ListMyTemplates(req *types.ListMyTemplatesReq) (resp *types.ListMyTemplatesResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.BotRpc.ListMyTemplates(l.ctx, &botclient.ListMyTemplatesReq{
		OwnerUid: uid,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.BotTemplateItem, 0, len(result.List))
	for _, t := range result.List {
		list = append(list, convertTemplate(t))
	}

	return &types.ListMyTemplatesResp{
		List: list,
	}, nil
}

func convertTemplate(t *botclient.BotTemplateInfo) types.BotTemplateItem {
	item := types.BotTemplateItem{
		TemplateId:        strconv.FormatInt(t.TemplateId, 10),
		Name:              t.Name,
		Avatar:            t.Avatar,
		Description:       t.Description,
		Category:          t.Category,
		Version:           t.Version,
		SystemPrompt:      t.SystemPrompt,
		WelcomeMessage:    t.WelcomeMessage,
		ConversationStyle: t.ConversationStyle,
		ToolDefinitions:   t.ToolDefinitions,
		KbStructure:       t.KbStructure,
		ConfigSchema:      t.ConfigSchema,
		SupportedModels:   t.SupportedModels,
		IsOfficial:        t.IsOfficial,
		Status:            t.Status,
		CreatedAt:         t.CreatedAt,
	}
	if t.Tags != nil {
		item.Tags = t.Tags
	}
	return item
}
