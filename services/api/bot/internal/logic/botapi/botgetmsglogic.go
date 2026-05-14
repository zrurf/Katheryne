package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

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
	return &types.BotGetMsgResp{
		MsgID:       req.MsgID,
		ConvID:      req.ConvID,
		SenderUID:   0,
		SenderName:  "",
		MsgType:     "text",
		Content:     "",
		ContentType: "text/plain",
		CreatedAt:   0,
	}, nil
}