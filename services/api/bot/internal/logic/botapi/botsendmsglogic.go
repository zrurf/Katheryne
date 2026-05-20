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
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if !l.svcCtx.InstallationDao.IsInstalled(l.ctx, auth.BotID, req.ConvID) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	msgID := time.Now().UnixNano()
	return &types.BotSendMsgResp{
		MsgID:     msgID,
		CreatedAt: time.Now().Unix(),
	}, nil
}