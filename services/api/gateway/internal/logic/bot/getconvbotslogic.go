package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConvBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConvBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvBotsLogic {
	return &GetConvBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConvBotsLogic) GetConvBots(req *types.GetConvBotsReq) (resp *types.GetConvBotsResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	result, err := l.svcCtx.BotRpc.GetConvBots(l.ctx, &botclient.GetConvBotsReq{
		ConvId: convId,
	})
	if err != nil {
		l.Errorf("GetConvBots RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.InstalledBotItem, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, types.InstalledBotItem{
			BotId:       strconv.FormatInt(item.BotId, 10),
			Name:        item.Name,
			Avatar:      item.Avatar,
			Description: item.Description,
			Permissions: item.Permissions,
			InstalledAt: item.InstalledAt,
		})
	}
	return &types.GetConvBotsResp{List: list}, nil
}
