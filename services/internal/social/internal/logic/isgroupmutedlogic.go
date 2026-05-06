package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsGroupMutedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsGroupMutedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsGroupMutedLogic {
	return &IsGroupMutedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsGroupMutedLogic) IsGroupMuted(in *social.IsGroupMutedReq) (*social.IsGroupMutedResp, error) {
	// todo: add your logic here and delete this line

	return &social.IsGroupMutedResp{}, nil
}
