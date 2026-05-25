package bot

import (
	"context"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotTemplateLogic {
	return &CreateBotTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateBotTemplateLogic) CreateBotTemplate(req *types.CreateBotTemplateReq) (resp *types.CreateBotTemplateResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.BotRpc.CreateBotTemplate(l.ctx, &botclient.CreateBotTemplateReq{
		OwnerUid:          uid,
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
	return &types.CreateBotTemplateResp{
		TemplateId: result.TemplateId,
	}, nil
}