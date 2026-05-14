package botapi

import (
	"context"
	"time"

	"bot/internal/svc"
	"bot/internal/types"

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
	msgID := time.Now().UnixNano()
	return &types.BotSendMsgResp{
		MsgID:     msgID,
		CreatedAt: time.Now().Unix(),
	}, nil
}