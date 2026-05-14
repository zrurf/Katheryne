package logic

import (
	"context"

	"conversation/internal/svc"
	"conversation/conversation"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLastMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLastMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLastMessageLogic {
	return &UpdateLastMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateLastMessageLogic) UpdateLastMessage(in *conversation.UpdateLastMessageReq) (*conversation.UpdateLastMessageResp, error) {
	err := l.svcCtx.ConversationDao.UpdateLastMessage(l.ctx, in.ConvId, in.MsgId, in.Snippet, in.Sender)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	return &conversation.UpdateLastMessageResp{}, nil
}