package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InstallBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InstallBotLogic {
	return &InstallBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InstallBotLogic) InstallBot(req *types.InstallBotReq) (resp *types.InstallBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}

	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	templateId, _ := strconv.ParseInt(req.TemplateId, 10, 64)

	result, err := l.svcCtx.BotRpc.InstallBot(l.ctx, &botclient.InstallBotReq{
		BotId:         botId,
		ConvId:        convId,
		Permissions:   perms,
		Uid:           uid,
		TemplateId:    templateId,
		ModelProvider: req.ModelProvider,
		ModelName:     req.ModelName,
		ApiKey:        req.ApiKey,
		KbConfig:      req.KbConfig,
	})
	if err != nil {
		return nil, err
	}

	instanceId := ""
	if result.InstanceId > 0 {
		instanceId = strconv.FormatInt(result.InstanceId, 10)
	}

	return &types.InstallBotResp{InstanceId: instanceId}, nil
}
