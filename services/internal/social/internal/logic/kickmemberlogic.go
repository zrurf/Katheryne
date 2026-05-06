package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type KickMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewKickMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KickMemberLogic {
	return &KickMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *KickMemberLogic) KickMember(in *social.KickMemberReq) (*social.KickMemberResp, error) {
	// todo: add your logic here and delete this line

	return &social.KickMemberResp{}, nil
}
