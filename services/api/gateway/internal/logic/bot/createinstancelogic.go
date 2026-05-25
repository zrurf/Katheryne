package bot

import (
	"context"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotInstanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotInstanceLogic {
	return &CreateBotInstanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateBotInstanceLogic) CreateBotInstance(req *types.CreateBotInstanceReq) (resp *types.CreateBotInstanceResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.BotRpc.CreateBotInstance(l.ctx, &botclient.CreateBotInstanceReq{
		Uid:            uid,
		TemplateId:     req.TemplateId,
		Name:           req.Name,
		Avatar:         req.Avatar,
		IsSelfHosted:   req.IsSelfHosted,
		HostedBy:       req.HostedBy,
		ModelProvider:  req.ModelProvider,
		ModelName:      req.ModelName,
		ApiKey:         req.ApiKey,
		ApiBaseUrl:     req.ApiBaseUrl,
		KbConfig:       req.KbConfig,
		InstanceConfig: req.InstanceConfig,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateBotInstanceResp{
		InstanceId: result.InstanceId,
		BotId:      result.BotId,
	}, nil
}