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
	// todo: add your logic here and delete this line

	return &message.SubmitReadResp{}, nil
}
