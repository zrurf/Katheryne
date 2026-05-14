package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type ClearUnreadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearUnreadLogic {
	return &ClearUnreadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearUnreadLogic) ClearUnread(in *conversation.ClearUnreadReq) (*conversation.ClearUnreadResp, error) {
	member, err := l.svcCtx.ConversationDao.GetConvMember(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not a member of this conversation")
		}
		l.Logger.Error(err)
		return nil, err
	}
	if !member.IsActive {
		return nil, errors.New("not an active member of this conversation")
	}

	err = l.svcCtx.ConversationDao.ClearUnread(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	_ = l.svcCtx.RedisDao.DelUnreadCache(l.ctx, in.Uid, in.ConvId)

	return &conversation.ClearUnreadResp{}, nil
}
