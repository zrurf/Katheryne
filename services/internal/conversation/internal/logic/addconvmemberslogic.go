package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type AddConvMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddConvMembersLogic {
	return &AddConvMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddConvMembersLogic) AddConvMembers(in *conversation.AddConvMembersReq) (*conversation.AddConvMembersResp, error) {
	conv, err := l.svcCtx.ConversationDao.GetConversationById(l.ctx, in.ConvId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("conversation not found")
		}
		l.Logger.Error(err)
		return nil, err
	}

	if conv.Type != "GROUP" {
		return nil, errors.New("can only add members to group conversation")
	}

	if len(in.Uids) > 0 {
		err = l.svcCtx.ConversationDao.BatchAddConvMembers(l.ctx, in.ConvId, in.Uids)
		if err != nil {
			l.Logger.Error(err)
			return nil, err
		}
		_ = l.svcCtx.RedisDao.DelConvListCaches(l.ctx, in.Uids)
	}

	_ = l.svcCtx.RedisDao.DelConvMembersCache(l.ctx, in.ConvId)

	return &conversation.AddConvMembersResp{}, nil
}
