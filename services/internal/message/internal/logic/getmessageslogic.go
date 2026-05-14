package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMessagesLogic) GetMessages(in *message.GetMessagesReq) (*message.GetMessagesResp, error) {
	l.Logger.Infof("GetMessages called: convId=%d, cursor=%d, limit=%d, direction=%s", in.ConvId, in.Cursor, in.Limit, in.Direction)
	var list []*message.MsgItem
	var err error

	if in.Direction == "after" {
		l.Logger.Infof("GetMessages using GetMessagesAfter: convId=%d, cursor=%d, limit=%d", in.ConvId, in.Cursor, in.Limit)
		msgs, err := l.svcCtx.MessageDao.GetMessagesAfter(l.ctx, in.ConvId, in.Cursor, in.Limit)
		if err != nil {
			l.Logger.Errorf("GetMessagesAfter failed: convId=%d, err=%v", in.ConvId, err)
			return nil, err
		}
		l.Logger.Infof("GetMessagesAfter returned %d messages for convId=%d", len(msgs), in.ConvId)
		list = make([]*message.MsgItem, len(msgs))
		for i, m := range msgs {
			list[i] = toMsgItem(m)
		}
		return &message.GetMessagesResp{List: list, HasMore: len(list) >= int(in.Limit)}, nil
	}

	if in.Cursor <= 0 {
		l.Logger.Infof("GetMessages using GetLatestMessages: convId=%d, limit=%d", in.ConvId, in.Limit)
		msgs, err := l.svcCtx.MessageDao.GetLatestMessages(l.ctx, in.ConvId, in.Limit)
		if err != nil {
			l.Logger.Errorf("GetLatestMessages failed: convId=%d, err=%v", in.ConvId, err)
			return nil, err
		}
		l.Logger.Infof("GetLatestMessages returned %d messages for convId=%d", len(msgs), in.ConvId)
		list = make([]*message.MsgItem, len(msgs))
		for i, m := range msgs {
			list[len(msgs)-1-i] = toMsgItem(m)
		}
		return &message.GetMessagesResp{List: list, HasMore: len(list) >= int(in.Limit)}, nil
	}

	l.Logger.Infof("GetMessages using GetMessagesBefore: convId=%d, cursor=%d, limit=%d", in.ConvId, in.Cursor, in.Limit)
	msgs, err := l.svcCtx.MessageDao.GetMessagesBefore(l.ctx, in.ConvId, in.Cursor, in.Limit)
	if err != nil {
		l.Logger.Errorf("GetMessagesBefore failed: convId=%d, err=%v", in.ConvId, err)
		return nil, err
	}
	l.Logger.Infof("GetMessagesBefore returned %d messages for convId=%d", len(msgs), in.ConvId)
	list = make([]*message.MsgItem, len(msgs))
	for i, m := range msgs {
		list[len(msgs)-1-i] = toMsgItem(m)
	}
	return &message.GetMessagesResp{List: list, HasMore: len(list) >= int(in.Limit)}, nil
}
