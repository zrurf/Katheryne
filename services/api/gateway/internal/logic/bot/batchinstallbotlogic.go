package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchInstallBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchInstallBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchInstallBotLogic {
	return &BatchInstallBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchInstallBotLogic) BatchInstallBot(req *types.BatchInstallBotReq) (resp *types.BatchInstallBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, err := strconv.ParseInt(req.BotId, 10, 64)
	if err != nil {
		return nil, err
	}
	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}

	var successCount int32
	var failedConvs []string

	for _, convIdStr := range req.ConvIds {
		convId, parseErr := strconv.ParseInt(convIdStr, 10, 64)
		if parseErr != nil {
			failedConvs = append(failedConvs, convIdStr)
			continue
		}
		_, installErr := l.svcCtx.BotRpc.InstallBot(l.ctx, &botclient.InstallBotReq{
			BotId:       botId,
			ConvId:      convId,
			Permissions: perms,
			Uid:         uid,
		})
		if installErr != nil {
			l.Errorf("BatchInstallBot: install bot %d to conv %d failed: %v", botId, convId, installErr)
			failedConvs = append(failedConvs, convIdStr)
		} else {
			successCount++
		}
	}

	return &types.BatchInstallBotResp{
		SuccessCount: successCount,
		FailedConvs:  failedConvs,
	}, nil
}
