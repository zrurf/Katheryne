package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBotRateLimitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBotRateLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBotRateLimitLogic {
	return &UpdateBotRateLimitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBotRateLimitLogic) UpdateBotRateLimit(req *types.UpdateBotRateLimitReq) (resp *types.UpdateBotRateLimitResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	_, err = l.svcCtx.BotRpc.UpdateBotRateLimit(l.ctx, &botclient.UpdateBotRateLimitReq{
		Uid:               uid,
		BotId:             botId,
		MessagesPerMinute: int64(req.MaxConcurrency),
		MessagesPerDay:    req.MaxRequests,
		ApiCallsPerMinute: int64(req.WindowSeconds),
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateBotRateLimitResp{}, nil
}
