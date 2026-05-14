package installation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bot/internal/svc"
	"bot/internal/types"

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
	key := fmt.Sprintf("bot_installations:%d", req.BotID)

	var list []types.BotInstallationItem
	data, err := l.svcCtx.Redis.Get(l.ctx, key).Result()
	if err == nil {
		json.Unmarshal([]byte(data), &list)
	}

	list = append(list, types.BotInstallationItem{
		ConvID:      req.ConvID,
		Permissions: req.Permissions,
		InstalledAt: time.Now().Unix(),
	})

	data2, _ := json.Marshal(list)
	l.svcCtx.Redis.Set(l.ctx, key, data2, 0)

	convKey := fmt.Sprintf("conv_bots:%d", req.ConvID)
	l.svcCtx.Redis.SAdd(l.ctx, convKey, req.BotID)

	return &types.InstallBotResp{}, nil
}