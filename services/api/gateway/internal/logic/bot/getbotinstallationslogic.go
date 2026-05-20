package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotInstallationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotInstallationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotInstallationsLogic {
	return &GetBotInstallationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotInstallationsLogic) GetBotInstallations(req *types.GetBotInstallationsReq) (resp *types.GetBotInstallationsResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	result, err := l.svcCtx.BotRpc.GetBotInstallations(l.ctx, &botclient.GetBotInstallationsReq{
		BotId: botId,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.InstallationItem, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, types.InstallationItem{
			ConvId:      strconv.FormatInt(item.ConvId, 10),
			ConvType:    item.ConvType,
			Permissions: item.Permissions,
			InstalledAt: item.InstalledAt,
		})
	}
	return &types.GetBotInstallationsResp{List: list}, nil
}
