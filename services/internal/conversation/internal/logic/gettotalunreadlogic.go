package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTotalUnreadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTotalUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTotalUnreadLogic {
	return &GetTotalUnreadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTotalUnreadLogic) GetTotalUnread(in *conversation.GetTotalUnreadReq) (*conversation.GetTotalUnreadResp, error) {
	count, err := l.svcCtx.ConversationDao.CountUnreadByUid(l.ctx, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &conversation.GetTotalUnreadResp{Count: count}, nil
}
