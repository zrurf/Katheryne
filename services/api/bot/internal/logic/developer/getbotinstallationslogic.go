package developer

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/svc"
	"bot/internal/types"

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
	data, err := l.svcCtx.Redis.Get(l.ctx, fmt.Sprintf("bot_installations:%d", req.BotID)).Result()
	if err != nil {
		return &types.GetBotInstallationsResp{List: []types.BotInstallationItem{}}, nil
	}

	var list []types.BotInstallationItem
	json.Unmarshal([]byte(data), &list)
	return &types.GetBotInstallationsResp{List: list}, nil
}