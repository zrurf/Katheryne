package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotSendMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotSendMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotSendMsgLogic {
	return &BotSendMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotSendMsgLogic) BotSendMsg(req *types.BotSendMsgReq) (resp *types.BotSendMsgResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	result, err := l.svcCtx.BotRpc.BotSendMsg(l.ctx, &botclient.BotSendMsgReq{
		ConvId:      convId,
		MsgType:     req.ContentType,
		Content:     req.Content,
		ContentType: req.ContentType,
	})
	if err != nil {
		return nil, err
	}
	return &types.BotSendMsgResp{
		MsgId:     strconv.FormatInt(result.MsgId, 10),
		CreatedAt: result.CreatedAt,
	}, nil
}
