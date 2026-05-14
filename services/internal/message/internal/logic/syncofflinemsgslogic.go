package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncOfflineMsgsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncOfflineMsgsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncOfflineMsgsLogic {
	return &SyncOfflineMsgsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SyncOfflineMsgsLogic) SyncOfflineMsgs(in *message.SyncOfflineMsgsReq) (*message.SyncOfflineMsgsResp, error) {
	msgs, err := l.svcCtx.MessageDao.SyncOfflineMessages(l.ctx, in.Uid, in.LastSyncMsgId, in.Limit, in.ConvIds)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	list := make([]*message.MsgItem, len(msgs))
	var maxId int64
	for i, m := range msgs {
		list[i] = toMsgItem(m)
		if m.Id > maxId {
			maxId = m.Id
		}
	}

	return &message.SyncOfflineMsgsResp{
		Messages:         list,
		HasMore:          len(list) >= int(in.Limit),
		NewLastSyncMsgId: maxId,
	}, nil
}
