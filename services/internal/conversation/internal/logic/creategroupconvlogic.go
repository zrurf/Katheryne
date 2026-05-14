package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
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
	existing, err := l.svcCtx.ConversationDao.GetGroupConversationByGroupId(l.ctx, in.GroupId)
	if err == nil && existing != nil {
		return &conversation.CreateGroupConvResp{ConvId: existing.ConvId}, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		l.Logger.Error(err)
		return nil, err
	}

	convId, err := l.svcCtx.ConversationDao.CreateGroupConversation(l.ctx, in.GroupId, in.Name, in.Avatar)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if len(in.MemberUids) > 0 {
		err = l.svcCtx.ConversationDao.BatchAddConvMembers(l.ctx, convId, in.MemberUids)
		if err != nil {
			l.Logger.Error(err)
		}
		_ = l.svcCtx.RedisDao.DelConvListCaches(l.ctx, in.MemberUids)
	}

	_ = l.svcCtx.RedisDao.DelConvMembersCache(l.ctx, convId)

	return &conversation.CreateGroupConvResp{ConvId: convId}, nil
}
