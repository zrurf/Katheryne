package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
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
	conv, err := l.svcCtx.ConversationDao.GetConversationById(l.ctx, in.ConvId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("conversation not found")
		}
		l.Logger.Error(err)
		return nil, err
	}

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

	return &conversation.GetConversationResp{
		ConvId:  conv.ConvId,
		Type:    conv.Type,
		Name:    nullString(conv.Name),
		Avatar:  nullString(conv.Avatar),
		GroupId: nullInt64(conv.GroupId),
		Mute:    member.Mute,
		Pinned:  member.Pinned,
		Uid:     nullInt64(conv.Uid),
		PeerUid: nullInt64(conv.PeerUid),
	}, nil
}
