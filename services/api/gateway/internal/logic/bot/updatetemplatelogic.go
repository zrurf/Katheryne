package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotTemplateLogic {
	return &UpdateBotTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBotTemplateLogic) UpdateBotTemplate(req *types.UpdateBotTemplateReq) (resp *types.UpdateBotTemplateResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	templateId, _ := strconv.ParseInt(req.TemplateId, 10, 64)
	_, err = l.svcCtx.BotRpc.UpdateBotTemplate(l.ctx, &botclient.UpdateBotTemplateReq{
		Uid:               uid,
		TemplateId:        templateId,
		Name:              req.Name,
		Avatar:            req.Avatar,
		Description:       req.Description,
		Category:          req.Category,
		SystemPrompt:      req.SystemPrompt,
		WelcomeMessage:    req.WelcomeMessage,
		ConversationStyle: req.ConversationStyle,
		ToolDefinitions:   req.ToolDefinitions,
		KbStructure:       req.KbStructure,
		ConfigSchema:      req.ConfigSchema,
		SupportedModels:   req.SupportedModels,
		Tags:              req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateBotTemplateResp{}, nil
}
