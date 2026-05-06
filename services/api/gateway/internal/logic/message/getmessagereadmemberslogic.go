// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package message

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessageReadMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMessageReadMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessageReadMembersLogic {
	return &GetMessageReadMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessageReadMembersLogic) GetMessageReadMembers(req *types.GetMessageReadMembersReq) (resp *types.GetMessageReadMembersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
