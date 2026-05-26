package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBotInstanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteBotInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBotInstanceLogic {
	return &DeleteBotInstanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBotInstanceLogic) DeleteBotInstance(req *types.DeleteBotInstanceReq) (resp *types.DeleteBotInstanceResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	instanceId, _ := strconv.ParseInt(req.InstanceId, 10, 64)
	_, err = l.svcCtx.BotRpc.DeleteBotInstance(l.ctx, &botclient.DeleteBotInstanceReq{
		Uid:        uid,
		InstanceId: instanceId,
	})
	if err != nil {
		return nil, err
	}
	return &types.DeleteBotInstanceResp{}, nil
}
