package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGroupConvInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGroupConvInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupConvInfoLogic {
	return &UpdateGroupConvInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateGroupConvInfoLogic) UpdateGroupConvInfo(in *conversation.UpdateGroupConvInfoReq) (*conversation.UpdateGroupConvInfoResp, error) {
	return &conversation.UpdateGroupConvInfoResp{}, nil
}
