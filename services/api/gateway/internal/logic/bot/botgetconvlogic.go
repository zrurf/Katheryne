package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvLogic {
	return &BotGetConvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetConvLogic) BotGetConv(req *types.BotGetConvReq) (resp *types.BotGetConvResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	result, err := l.svcCtx.BotRpc.BotGetConv(l.ctx, &botclient.BotGetConvReq{
		ConvId: convId,
	})
	if err != nil {
		return nil, err
	}
	return &types.BotGetConvResp{
		ConvId:    strconv.FormatInt(result.ConvId, 10),
		Type:      result.Type,
		Name:      result.Name,
		Avatar:    result.Avatar,
		GroupId:   result.GroupId,
		CreatedAt: result.CreatedAt,
	}, nil
}
