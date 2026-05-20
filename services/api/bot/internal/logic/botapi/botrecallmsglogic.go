package botapi

import (
	"context"
	"fmt"

	"bot/internal/middleware"
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
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if !l.svcCtx.InstallationDao.IsInstalled(l.ctx, auth.BotID, req.ConvID) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	return &types.BotRecallMsgResp{}, nil
}