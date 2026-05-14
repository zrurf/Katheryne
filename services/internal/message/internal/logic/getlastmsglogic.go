package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLastMsgLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLastMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLastMsgLogic {
	return &GetLastMsgLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLastMsgLogic) GetLastMsg(in *message.GetLastMsgReq) (*message.GetLastMsgResp, error) {
	m, err := l.svcCtx.MessageDao.GetLastMessage(l.ctx, in.ConvId)
	if err != nil {
		l.Logger.Error(err)
		return &message.GetLastMsgResp{}, nil
	}

	return &message.GetLastMsgResp{
		Msg: toMsgItem(m),
	}, nil
}
