package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBotTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteBotTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBotTemplateLogic {
	return &DeleteBotTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBotTemplateLogic) DeleteBotTemplate(req *types.DeleteBotTemplateReq) (resp *types.DeleteBotTemplateResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	templateId, _ := strconv.ParseInt(req.TemplateId, 10, 64)
	_, err = l.svcCtx.BotRpc.DeleteBotTemplate(l.ctx, &botclient.DeleteBotTemplateReq{
		Uid:        uid,
		TemplateId: templateId,
	})
	if err != nil {
		return nil, err
	}
	return &types.DeleteBotTemplateResp{}, nil
}
