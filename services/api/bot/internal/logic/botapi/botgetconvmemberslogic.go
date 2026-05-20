package botapi

import (
	"context"
	"fmt"

	"bot/internal/middleware"
	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvMembersLogic {
	return &BotGetConvMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetConvMembersLogic) BotGetConvMembers(req *types.BotGetConvMembersReq) (resp *types.BotGetConvMembersResp, err error) {
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if !l.svcCtx.InstallationDao.IsInstalled(l.ctx, auth.BotID, req.ConvID) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	members, err := l.svcCtx.InstallationDao.GetGroupMembers(l.ctx, req.ConvID)
	if err != nil {
		return nil, err
	}

	return &types.BotGetConvMembersResp{Members: members}, nil
}