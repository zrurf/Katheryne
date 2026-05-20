package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RotateWebhookSecretLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRotateWebhookSecretLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RotateWebhookSecretLogic {
	return &RotateWebhookSecretLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RotateWebhookSecretLogic) RotateWebhookSecret(req *types.RotateWebhookSecretReq) (resp *types.RotateWebhookSecretResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	result, err := l.svcCtx.BotRpc.RotateWebhookSecret(l.ctx, &botclient.RotateWebhookSecretReq{
		BotId: botId,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.RotateWebhookSecretResp{
		WebhookSecret: result.WebhookSecret,
	}, nil
}
