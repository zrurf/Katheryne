package bot

import (
	"context"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyInstancesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyInstancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyInstancesLogic {
	return &ListMyInstancesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyInstancesLogic) ListMyInstances(req *types.ListMyInstancesReq) (resp *types.ListMyInstancesResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.BotRpc.ListMyInstances(l.ctx, &botclient.ListMyInstancesReq{
		OwnerUid: uid,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.BotInstanceItem, 0, len(result.List))
	for _, inst := range result.List {
		list = append(list, convertInstance(inst))
	}

	return &types.ListMyInstancesResp{
		List: list,
	}, nil
}

func convertInstance(inst *botclient.BotInstanceInfo) types.BotInstanceItem {
	return types.BotInstanceItem{
		InstanceId:    inst.InstanceId,
		BotId:         inst.BotId,
		TemplateId:    inst.TemplateId,
		Name:          inst.Name,
		Avatar:        inst.Avatar,
		IsSelfHosted:  inst.IsSelfHosted,
		HostedBy:      inst.HostedBy,
		ModelProvider: inst.ModelProvider,
		ModelName:     inst.ModelName,
		KbConfig:      inst.KbConfig,
		Status:        inst.Status,
		CreatedAt:     inst.CreatedAt,
	}
}