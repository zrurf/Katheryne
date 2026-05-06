package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupConvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupConvLogic {
	return &CreateGroupConvLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateGroupConvLogic) CreateGroupConv(in *conversation.CreateGroupConvReq) (*conversation.CreateGroupConvResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.CreateGroupConvResp{}, nil
}
