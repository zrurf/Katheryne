package logic

import (
	"context"
	"errors"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecallMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecallMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecallMessageLogic {
	return &RecallMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RecallMessageLogic) RecallMessage(in *message.RecallMessageReq) (*message.RecallMessageResp, error) {
	m, err := l.svcCtx.MessageDao.GetMessageById(l.ctx, in.MsgId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if m.Sender != in.Operator {
		return nil, errors.New("只能撤回自己发送的消息")
	}

	if m.Recalled {
		return &message.RecallMessageResp{}, nil
	}

	err = l.svcCtx.MessageDao.RecallMessage(l.ctx, in.MsgId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelLastMsgCache(l.ctx, in.ConvId)
	if err != nil {
		l.Logger.Error("del last msg cache error:", err)
	}

	return &message.RecallMessageResp{}, nil
}
