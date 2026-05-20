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
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}
	_, err = l.svcCtx.BotRpc.InstallBot(l.ctx, &botclient.InstallBotReq{
		BotId:       botId,
		ConvId:      convId,
		Permissions: perms,
		Uid:         uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.InstallBotResp{}, nil
}
