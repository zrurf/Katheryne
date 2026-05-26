package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotTemplateLogic {
	return &GetBotTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotTemplateLogic) GetBotTemplate(req *types.GetBotTemplateReq) (resp *types.GetBotTemplateResp, err error) {
	templateId, _ := strconv.ParseInt(req.TemplateId, 10, 64)
	result, err := l.svcCtx.BotRpc.GetBotTemplate(l.ctx, &botclient.GetBotTemplateReq{
		TemplateId: templateId,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetBotTemplateResp{
		Template: convertTemplate(result.Template),
	}, nil
}
