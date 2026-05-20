package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetUserLogic {
	return &BotGetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetUserLogic) BotGetUser(req *types.BotGetUserReq) (resp *types.BotGetUserResp, err error) {
	uid, _ := strconv.ParseInt(req.Uid, 10, 64)
	result, err := l.svcCtx.BotRpc.BotGetUser(l.ctx, &botclient.BotGetUserReq{
		Uid: uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.BotGetUserResp{
		Uid:    strconv.FormatInt(result.Uid, 10),
		Name:   result.Name,
		Avatar: result.Avatar,
	}, nil
}
