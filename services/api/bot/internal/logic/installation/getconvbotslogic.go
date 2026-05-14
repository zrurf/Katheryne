package installation

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

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
	convKey := fmt.Sprintf("conv_bots:%d", req.ConvID)
	botIDs, err := l.svcCtx.Redis.SMembers(l.ctx, convKey).Result()
	if err != nil {
		return &types.GetConvBotsResp{List: []types.InstalledBotItem{}}, nil
	}

	var list []types.InstalledBotItem
	for _, id := range botIDs {
		data, err := l.svcCtx.Redis.HGet(l.ctx, "bots", id).Result()
		if err != nil {
			continue
		}
		var bot types.BotInfo
		json.Unmarshal([]byte(data), &bot)
		list = append(list, types.InstalledBotItem{
			BotID:       bot.BotID,
			Name:        bot.Name,
			Avatar:      bot.Avatar,
			Description: bot.Description,
		})
	}

	return &types.GetConvBotsResp{List: list}, nil
}