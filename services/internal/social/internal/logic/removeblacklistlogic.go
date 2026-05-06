package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveBlacklistLogic {
	return &RemoveBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveBlacklistLogic) RemoveBlacklist(in *social.RemoveBlacklistReq) (*social.RemoveBlacklistResp, error) {
	// todo: add your logic here and delete this line

	return &social.RemoveBlacklistResp{}, nil
}
