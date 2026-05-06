package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetConvPinLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetConvPinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConvPinLogic {
	return &SetConvPinLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetConvPinLogic) SetConvPin(in *conversation.SetConvPinReq) (*conversation.SetConvPinResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.SetConvPinResp{}, nil
}
