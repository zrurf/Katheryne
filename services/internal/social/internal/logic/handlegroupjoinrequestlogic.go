package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupJoinRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleGroupJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupJoinRequestLogic {
	return &HandleGroupJoinRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HandleGroupJoinRequestLogic) HandleGroupJoinRequest(in *social.HandleGroupJoinReq) (*social.HandleGroupJoinResp, error) {
	// todo: add your logic here and delete this line

	return &social.HandleGroupJoinResp{}, nil
}
