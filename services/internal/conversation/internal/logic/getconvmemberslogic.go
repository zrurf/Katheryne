package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConvMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvMembersLogic {
	return &GetConvMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConvMembersLogic) GetConvMembers(in *conversation.GetConvMembersReq) (*conversation.GetConvMembersResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.GetConvMembersResp{}, nil
}
