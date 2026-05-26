package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotInstanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotInstanceLogic {
	return &GetBotInstanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotInstanceLogic) GetBotInstance(req *types.GetBotInstanceReq) (resp *types.GetBotInstanceResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	instanceId, _ := strconv.ParseInt(req.InstanceId, 10, 64)
	result, err := l.svcCtx.BotRpc.GetBotInstance(l.ctx, &botclient.GetBotInstanceReq{
		InstanceId: instanceId,
		Uid:        uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetBotInstanceResp{
		Instance: convertInstance(result.Instance),
	}, nil
}
