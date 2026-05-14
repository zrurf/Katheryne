package installation

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UninstallBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUninstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UninstallBotLogic {
	return &UninstallBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UninstallBotLogic) UninstallBot(req *types.UninstallBotReq) (resp *types.UninstallBotResp, err error) {
	key := fmt.Sprintf("bot_installations:%d", req.BotID)

	data, err := l.svcCtx.Redis.Get(l.ctx, key).Result()
	if err != nil {
		return &types.UninstallBotResp{}, nil
	}

	var list []types.BotInstallationItem
	json.Unmarshal([]byte(data), &list)

	var newList []types.BotInstallationItem
	for _, item := range list {
		if item.ConvID != req.ConvID {
			newList = append(newList, item)
		}
	}

	data2, _ := json.Marshal(newList)
	l.svcCtx.Redis.Set(l.ctx, key, data2, 0)

	convKey := fmt.Sprintf("conv_bots:%d", req.ConvID)
	l.svcCtx.Redis.SRem(l.ctx, convKey, req.BotID)

	return &types.UninstallBotResp{}, nil
}