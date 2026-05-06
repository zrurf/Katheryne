package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveConvMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveConvMembersLogic {
	return &RemoveConvMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveConvMembersLogic) RemoveConvMembers(in *conversation.RemoveConvMembersReq) (*conversation.RemoveConvMembersResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.RemoveConvMembersResp{}, nil
}
