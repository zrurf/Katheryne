package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConversationsLogic) GetConversations(in *conversation.GetConversationsReq) (*conversation.GetConversationsResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.GetConversationsResp{}, nil
}
