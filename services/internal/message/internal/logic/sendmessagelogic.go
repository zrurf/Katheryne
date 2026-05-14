package logic

import (
	"context"

	"message/internal/dao"
	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendMessageLogic) SendMessage(in *message.SendMessageReq) (*message.SendMessageResp, error) {
	l.Logger.Infof("SendMessage called: convId=%d, sender=%d, receiver=%d, type=%s, content=%s", in.ConvId, in.Sender, in.Receiver, in.Type, in.Content)
	m, err := l.svcCtx.MessageDao.InsertMessage(l.ctx, in.ConvId, in.Sender, in.Receiver, in.Type, in.Content, in.ContentType, in.QuoteMsgId, in.Extra)
	if err != nil {
		l.Logger.Errorf("InsertMessage failed: convId=%d, err=%v", in.ConvId, err)
		return nil, err
	}
	l.Logger.Infof("InsertMessage success: msgId=%d, convId=%d", m.Id, m.ConvId)

	snippet := in.Content
	if len(snippet) > 50 {
		snippet = snippet[:50] + "..."
	}

	err = l.svcCtx.RedisDao.SetLastMsgCache(l.ctx, in.ConvId, m.Id, snippet, in.Sender, m.CreatedAt.UnixMilli())
	if err != nil {
		l.Logger.Error("set last msg cache error:", err)
	}

	return &message.SendMessageResp{
		MsgId:     m.Id,
		ConvId:    in.ConvId,
		CreatedAt: m.CreatedAt.UnixMilli(),
	}, nil
}

func toMsgItem(m *dao.Message) *message.MsgItem {
	item := &message.MsgItem{
		Id:          m.Id,
		ConvId:      m.ConvId,
		Sender:      m.Sender,
		Receiver:    m.Receiver,
		Type:        m.Type,
		Content:     m.Content,
		ContentType: m.ContentType,
		Recalled:    m.Recalled,
		Edited:      m.Edited,
		CreatedAt:   m.CreatedAt.UnixMilli(),
	}
	if m.QuoteMsgId.Valid {
		item.QuoteMsgId = m.QuoteMsgId.Int64
	}
	if m.Extra.Valid {
		item.Extra = m.Extra.String
	}
	return item
}
