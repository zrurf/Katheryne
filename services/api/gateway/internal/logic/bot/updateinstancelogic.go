package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotInstanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotInstanceLogic {
	return &UpdateBotInstanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBotInstanceLogic) UpdateBotInstance(req *types.UpdateBotInstanceReq) (resp *types.UpdateBotInstanceResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	instanceId, _ := strconv.ParseInt(req.InstanceId, 10, 64)
	_, err = l.svcCtx.BotRpc.UpdateBotInstance(l.ctx, &botclient.UpdateBotInstanceReq{
		Uid:            uid,
		InstanceId:     instanceId,
		Name:           req.Name,
		Avatar:         req.Avatar,
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
	return &types.UpdateBotInstanceResp{}, nil
}
