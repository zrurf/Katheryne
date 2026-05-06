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
	// todo: add your logic here and delete this line

	return &message.GetLastMsgResp{}, nil
}
