package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReadLogic {
	return &SubmitReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitReadLogic) SubmitRead(in *message.SubmitReadReq) (*message.SubmitReadResp, error) {
	startMsgId := in.StartMsgId
	endMsgId := in.EndMsgId
	if startMsgId <= 0 || endMsgId <= 0 {
		startMsgId = in.LastReadMsgId
		endMsgId = in.LastReadMsgId
	}

	err := l.svcCtx.MessageDao.SubmitReadInterval(l.ctx, in.ConvId, in.Uid, startMsgId, endMsgId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelUnreadCache(l.ctx, in.Uid, in.ConvId)
	if err != nil {
		l.Logger.Error("del unread cache error:", err)
	}

	return &message.SubmitReadResp{}, nil
}
