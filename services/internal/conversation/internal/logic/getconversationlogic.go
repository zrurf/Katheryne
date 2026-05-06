package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationLogic {
	return &GetConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConversationLogic) GetConversation(in *conversation.GetConversationReq) (*conversation.GetConversationResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.GetConversationResp{}, nil
}
