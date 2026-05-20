package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

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

func (l *ListMyBotsLogic) ListMyBots(req *types.ListMyBotsReq) (resp *types.ListMyBotsResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.BotRpc.ListMyBots(l.ctx, &botclient.ListMyBotsReq{
		OwnerUid: uid,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.BotItem, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, types.BotItem{
			BotId:       strconv.FormatInt(item.BotId, 10),
			Name:        item.Name,
			Description: item.Description,
			Avatar:      item.Avatar,
			Status:      item.Status,
			ClientId:    item.ClientId,
			CreatedAt:   item.CreatedAt,
		})
	}
	return &types.ListMyBotsResp{List: list, Total: int64(len(list))}, nil
}
