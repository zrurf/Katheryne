// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
