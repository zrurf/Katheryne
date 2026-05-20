package botapi

import (
	"context"
	"fmt"

	"bot/internal/middleware"
	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvLogic {
	return &BotGetConvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetConvLogic) BotGetConv(req *types.BotGetConvReq) (resp *types.BotGetConvResp, err error) {
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if !l.svcCtx.InstallationDao.IsInstalled(l.ctx, auth.BotID, req.ConvID) {
		return nil, fmt.Errorf("bot not installed in this conversation")
	}

	convType, name, avatar, groupID, createdAt, err := l.svcCtx.InstallationDao.GetConvInfo(l.ctx, req.ConvID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	resp = &types.BotGetConvResp{
		ConvID:    req.ConvID,
		Type:      convType,
		Name:      name,
		Avatar:    avatar,
		CreatedAt: createdAt,
	}
	if groupID > 0 {
		resp.GroupID = groupID
	}

	return resp, nil
}