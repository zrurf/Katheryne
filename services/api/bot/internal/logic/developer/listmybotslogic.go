package developer

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyBotsLogic {
	return &ListMyBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyBotsLogic) ListMyBots() (resp *types.ListMyBotsResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	botIDs, err := l.svcCtx.Redis.SMembers(l.ctx, fmt.Sprintf("user_bots:%d", uid)).Result()
	if err != nil {
		return &types.ListMyBotsResp{List: []types.BotInfo{}}, nil
	}

	var list []types.BotInfo
	for _, id := range botIDs {
		data, err := l.svcCtx.Redis.HGet(l.ctx, "bots", id).Result()
		if err != nil {
			continue
		}
		var bot types.BotInfo
		json.Unmarshal([]byte(data), &bot)
		list = append(list, bot)
	}

	return &types.ListMyBotsResp{List: list}, nil
}