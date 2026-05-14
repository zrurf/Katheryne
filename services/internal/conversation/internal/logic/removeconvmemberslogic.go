package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
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
	conv, err := l.svcCtx.ConversationDao.GetConversationById(l.ctx, in.ConvId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("conversation not found")
		}
		l.Logger.Error(err)
		return nil, err
	}

	if conv.Type != "GROUP" {
		return nil, errors.New("can only remove members from group conversation")
	}

	if len(in.Uids) > 0 {
		err = l.svcCtx.ConversationDao.BatchRemoveConvMembers(l.ctx, in.ConvId, in.Uids)
		if err != nil {
			l.Logger.Error(err)
			return nil, err
		}
		_ = l.svcCtx.RedisDao.DelConvListCaches(l.ctx, in.Uids)
	}

	_ = l.svcCtx.RedisDao.DelConvMembersCache(l.ctx, in.ConvId)

	return &conversation.RemoveConvMembersResp{}, nil
}
