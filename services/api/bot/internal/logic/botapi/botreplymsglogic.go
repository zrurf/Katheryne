package botapi

import (
	"context"
	"fmt"
	"time"

	"bot/internal/middleware"
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
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if !l.svcCtx.InstallationDao.IsInstalled(l.ctx, auth.BotID, req.ConvID) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	msgID := time.Now().UnixNano()
	return &types.BotReplyMsgResp{
		MsgID:     msgID,
		CreatedAt: time.Now().Unix(),
	}, nil
}