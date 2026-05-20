package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommunityBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCommunityBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommunityBotsLogic {
	return &ListCommunityBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCommunityBotsLogic) ListCommunityBots(req *types.ListCommunityBotsReq) (resp *types.ListCommunityBotsResp, err error) {
	result, err := l.svcCtx.BotRpc.ListCommunityBots(l.ctx, &botclient.ListCommunityBotsReq{
		Keyword: req.Keyword,
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
	return &types.ListCommunityBotsResp{List: list}, nil
}