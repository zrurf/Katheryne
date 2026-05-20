package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotRateLimitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotRateLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotRateLimitLogic {
	return &GetBotRateLimitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotRateLimitLogic) GetBotRateLimit(req *types.GetBotRateLimitReq) (resp *types.GetBotRateLimitResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	result, err := l.svcCtx.BotRpc.GetBotRateLimit(l.ctx, &botclient.GetBotRateLimitReq{
		BotId: botId,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetBotRateLimitResp{
		BotId:          req.BotId,
		MaxConcurrency: int32(result.MessagesPerMinute),
		MaxRequests:    result.MessagesPerDay,
		WindowSeconds:  int32(result.ApiCallsPerMinute),
		IsCustom:       true,
	}, nil
}
