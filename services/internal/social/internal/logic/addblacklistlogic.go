package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBlacklistLogic {
	return &AddBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 黑名单
func (l *AddBlacklistLogic) AddBlacklist(in *social.AddBlacklistReq) (*social.AddBlacklistResp, error) {
	// todo: add your logic here and delete this line

	return &social.AddBlacklistResp{}, nil
}
