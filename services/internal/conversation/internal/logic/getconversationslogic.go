package logic

import (
	"context"
	"database/sql"
	"time"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConversationsLogic) GetConversations(in *conversation.GetConversationsReq) (*conversation.GetConversationsResp, error) {
	convList, err := l.svcCtx.ConversationDao.ListConversationsByUid(l.ctx, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	items := make([]*conversation.ConversationItem, 0, len(convList))
	for _, c := range convList {
		member, err := l.svcCtx.ConversationDao.GetConvMember(l.ctx, c.ConvId, in.Uid)
		if err != nil {
			l.Logger.Error(err)
			continue
		}

		unread, err := l.svcCtx.ConversationDao.GetUnreadCount(l.ctx, c.ConvId, in.Uid)
		if err != nil {
			l.Logger.Error(err)
			unread = 0
		}

		var lastMsgTime int64
		if c.LastMsgTime.Valid {
			lastMsgTime = c.LastMsgTime.Time.UnixMilli()
		}

		item := &conversation.ConversationItem{
			ConvId:         c.ConvId,
			Type:           c.Type,
			Name:           nullString(c.Name),
			Avatar:         nullString(c.Avatar),
			GroupId:        nullInt64(c.GroupId),
			LastMsgId:      nullInt64(c.LastMsgId),
			LastMsgSnippet: nullString(c.LastMsgSnippet),
			LastMsgTime:    lastMsgTime,
			LastMsgSender:  nullInt64(c.LastMsgSender),
			UnreadCount:    unread,
			Mute:           member.Mute,
			Pinned:         member.Pinned,
			Uid:            nullInt64(c.Uid),
			PeerUid:        nullInt64(c.PeerUid),
		}
		items = append(items, item)
	}

	return &conversation.GetConversationsResp{List: items}, nil
}

func nullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullInt64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}

func nullTimeToMilli(t sql.NullTime) int64 {
	if t.Valid {
		return t.Time.UnixMilli()
	}
	return 0
}

func timeToMilli(t time.Time) int64 {
	return t.UnixMilli()
}
