package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetMsgLogic {
	return &BotGetMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetMsgLogic) BotGetMsg(req *types.BotGetMsgReq) (resp *types.BotGetMsgResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	msgId, _ := strconv.ParseInt(req.MsgId, 10, 64)
	result, err := l.svcCtx.BotRpc.BotGetMsg(l.ctx, &botclient.BotGetMsgReq{
		ConvId: convId,
		MsgId:  msgId,
	})
	if err != nil {
		return nil, err
	}
	return &types.BotGetMsgResp{
		Msg: &types.BotMsgItem{
			MsgId:       strconv.FormatInt(result.MsgId, 10),
			ConvId:      strconv.FormatInt(result.ConvId, 10),
			Sender:      strconv.FormatInt(result.SenderUid, 10),
			SenderName:  result.SenderName,
			Type:        result.MsgType,
			Content:     result.Content,
			ContentType: result.ContentType,
			Recalled:    result.Recalled,
			CreatedAt:   result.CreatedAt,
		},
	}, nil
}
