package developer

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotLogic {
	return &GetBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotLogic) GetBot(req *types.GetBotReq) (resp *types.GetBotResp, err error) {
	data, err := l.svcCtx.Redis.HGet(l.ctx, "bots", fmt.Sprintf("%d", req.BotID)).Result()
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	var bot types.BotInfo
	json.Unmarshal([]byte(data), &bot)
	return &types.GetBotResp{Bot: bot}, nil
}