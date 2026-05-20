package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotRecallMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotRecallMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotRecallMsgLogic {
	return &BotRecallMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotRecallMsgLogic) BotRecallMsg(req *types.BotRecallMsgReq) (resp *types.BotRecallMsgResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	msgId, _ := strconv.ParseInt(req.MsgId, 10, 64)
	_, err = l.svcCtx.BotRpc.BotRecallMsg(l.ctx, &botclient.BotRecallMsgReq{
		ConvId: convId,
		MsgId:  msgId,
	})
	if err != nil {
		return nil, err
	}
	return &types.BotRecallMsgResp{}, nil
}
