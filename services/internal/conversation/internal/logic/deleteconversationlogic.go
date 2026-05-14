package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConversationLogic {
	return &DeleteConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteConversationLogic) DeleteConversation(in *conversation.DeleteConversationReq) (*conversation.DeleteConversationResp, error) {
	conv, err := l.svcCtx.ConversationDao.GetConversationById(l.ctx, in.ConvId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &conversation.DeleteConversationResp{}, nil
		}
		l.Logger.Error(err)
		return nil, err
	}

	member, err := l.svcCtx.ConversationDao.GetConvMember(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &conversation.DeleteConversationResp{}, nil
		}
		l.Logger.Error(err)
		return nil, err
	}
	if !member.IsActive {
		return &conversation.DeleteConversationResp{}, nil
	}

	if conv.Type == "SINGLE" {
		_ = l.svcCtx.ConversationDao.SetConvMemberActive(l.ctx, in.ConvId, in.Uid, false)
	} else {
		_ = l.svcCtx.ConversationDao.SetConvMemberActive(l.ctx, in.ConvId, in.Uid, false)
	}

	_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, in.Uid)
	_ = l.svcCtx.RedisDao.DelConvMembersCache(l.ctx, in.ConvId)
	_ = l.svcCtx.RedisDao.DelConvCache(l.ctx, in.ConvId)

	return &conversation.DeleteConversationResp{}, nil
}
