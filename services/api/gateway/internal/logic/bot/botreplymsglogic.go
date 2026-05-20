package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotReplyMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotReplyMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotReplyMsgLogic {
	return &BotReplyMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotReplyMsgLogic) BotReplyMsg(req *types.BotReplyMsgReq) (resp *types.BotReplyMsgResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	msgId, _ := strconv.ParseInt(req.MsgId, 10, 64)
	result, err := l.svcCtx.BotRpc.BotReplyMsg(l.ctx, &botclient.BotReplyMsgReq{
		ConvId:       convId,
		ReplyToMsgId: msgId,
		MsgType:      req.ContentType,
		Content:      req.Content,
		ContentType:  req.ContentType,
	})
	if err != nil {
		return nil, err
	}
	return &types.BotReplyMsgResp{
		MsgId:     strconv.FormatInt(result.MsgId, 10),
		CreatedAt: result.CreatedAt,
	}, nil
}
