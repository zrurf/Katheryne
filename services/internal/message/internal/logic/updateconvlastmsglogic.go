package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateConvLastMsgLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateConvLastMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConvLastMsgLogic {
	return &UpdateConvLastMsgLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateConvLastMsgLogic) UpdateConvLastMsg(in *message.UpdateConvLastMsgReq) (*message.UpdateConvLastMsgResp, error) {
	// todo: add your logic here and delete this line

	return &message.UpdateConvLastMsgResp{}, nil
}
