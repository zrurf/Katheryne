package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncrementUnreadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrementUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrementUnreadLogic {
	return &IncrementUnreadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IncrementUnreadLogic) IncrementUnread(in *conversation.IncrementUnreadReq) (*conversation.IncrementUnreadResp, error) {
	err := l.svcCtx.ConversationDao.IncrementUnread(l.ctx, in.ConvId, in.Uids)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &conversation.IncrementUnreadResp{}, nil
}
