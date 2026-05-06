package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlacklistLogic {
	return &GetBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBlacklistLogic) GetBlacklist(in *social.GetBlacklistReq) (*social.GetBlacklistResp, error) {
	// todo: add your logic here and delete this line

	return &social.GetBlacklistResp{}, nil
}
