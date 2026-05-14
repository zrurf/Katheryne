package botapi

import (
	"context"
	"time"

	"bot/internal/svc"
	"bot/internal/types"

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
	msgID := time.Now().UnixNano()
	return &types.BotReplyMsgResp{
		MsgID:     msgID,
		CreatedAt: time.Now().Unix(),
	}, nil
}