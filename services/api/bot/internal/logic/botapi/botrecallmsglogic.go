package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

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
	return &types.BotRecallMsgResp{}, nil
}