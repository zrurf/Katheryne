package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOnlineStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOnlineStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOnlineStatusLogic {
	return &GetOnlineStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOnlineStatusLogic) GetOnlineStatus(in *social.GetOnlineStatusReq) (*social.GetOnlineStatusResp, error) {
	user, err := l.svcCtx.UserDBDao.GetUserById(l.ctx, in.Uid)
	if err != nil {
		return &social.GetOnlineStatusResp{Status: "offline"}, nil
	}
	return &social.GetOnlineStatusResp{Status: user.Status}, nil
}
