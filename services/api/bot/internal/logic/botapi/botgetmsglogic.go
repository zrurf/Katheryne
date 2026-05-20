package botapi

import (
	"context"
	"fmt"

	"bot/internal/middleware"
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
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if !l.svcCtx.InstallationDao.IsInstalled(l.ctx, auth.BotID, req.ConvID) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	return l.svcCtx.InstallationDao.GetMessage(l.ctx, req.MsgID, req.ConvID)
}
