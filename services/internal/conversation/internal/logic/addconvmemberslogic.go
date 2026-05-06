package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddConvMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddConvMembersLogic {
	return &AddConvMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddConvMembersLogic) AddConvMembers(in *conversation.AddConvMembersReq) (*conversation.AddConvMembersResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.AddConvMembersResp{}, nil
}
