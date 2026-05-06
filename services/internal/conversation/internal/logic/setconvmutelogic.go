package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetConvMuteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetConvMuteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConvMuteLogic {
	return &SetConvMuteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetConvMuteLogic) SetConvMute(in *conversation.SetConvMuteReq) (*conversation.SetConvMuteResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.SetConvMuteResp{}, nil
}
