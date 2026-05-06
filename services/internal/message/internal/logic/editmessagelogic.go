package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &message.EditMessageResp{}, nil
}
