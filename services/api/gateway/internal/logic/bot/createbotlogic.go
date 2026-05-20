package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBotLogic {
	return &CreateBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateBotLogic) CreateBot(req *types.CreateBotReq) (resp *types.CreateBotResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.BotRpc.CreateBot(l.ctx, &botclient.CreateBotReq{
		OwnerUid:    uid,
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		WebhookUrl:  req.WebhookUrl,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateBotResp{
		BotId:        strconv.FormatInt(result.BotId, 10),
		ClientId:     result.ClientId,
		ClientSecret: result.ClientSecret,
	}, nil
}
