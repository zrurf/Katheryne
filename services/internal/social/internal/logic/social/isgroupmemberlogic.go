package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsGroupMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsGroupMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsGroupMemberLogic {
	return &IsGroupMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsGroupMemberLogic) IsGroupMember(in *social.IsGroupMemberReq) (*social.IsGroupMemberResp, error) {
	member, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, in.GroupId, in.Uid)
	if err != nil || member == nil {
		return &social.IsGroupMemberResp{IsMember: false}, nil
	}
	return &social.IsGroupMemberResp{IsMember: true}, nil
}