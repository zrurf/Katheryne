package logic

import (
	"context"
	"errors"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type EditMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEditMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EditMessageLogic {
	return &EditMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EditMessageLogic) EditMessage(in *message.EditMessageReq) (*message.EditMessageResp, error) {
	m, err := l.svcCtx.MessageDao.GetMessageById(l.ctx, in.MsgId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if m.Sender != in.Editor {
		return nil, errors.New("只能编辑自己发送的消息")
	}

	if m.Recalled {
		return nil, errors.New("已撤回的消息不能编辑")
	}

	err = l.svcCtx.MessageDao.EditMessage(l.ctx, in.MsgId, in.Content)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelLastMsgCache(l.ctx, in.ConvId)
	if err != nil {
		l.Logger.Error("del last msg cache error:", err)
	}

	return &message.EditMessageResp{}, nil
}
