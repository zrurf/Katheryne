package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegenerateCredentialLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegenerateCredentialLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegenerateCredentialLogic {
	return &RegenerateCredentialLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegenerateCredentialLogic) RegenerateCredential(req *types.RegenerateCredentialReq) (resp *types.RegenerateCredentialResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	botId, _ := strconv.ParseInt(req.BotId, 10, 64)
	result, err := l.svcCtx.BotRpc.RegenerateCredential(l.ctx, &botclient.RegenerateCredentialReq{
		BotId: botId,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.RegenerateCredentialResp{
		ClientId:     result.ClientId,
		ClientSecret: result.ClientSecret,
	}, nil
}
