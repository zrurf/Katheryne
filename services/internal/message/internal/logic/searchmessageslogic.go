package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMessagesLogic {
	return &SearchMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchMessagesLogic) SearchMessages(in *message.SearchMessagesReq) (*message.SearchMessagesResp, error) {
	msgs, total, err := l.svcCtx.MessageDao.SearchMessages(l.ctx, in.Keyword, in.ConvId, in.Sender, in.StartTime, in.EndTime, in.Page, in.Size)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	list := make([]*message.MsgItem, len(msgs))
	for i, m := range msgs {
		list[i] = toMsgItem(m)
	}

	return &message.SearchMessagesResp{
		List:  list,
		Total: total,
	}, nil
}
