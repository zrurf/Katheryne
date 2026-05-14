package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFriendRemarkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFriendRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFriendRemarkLogic {
	return &UpdateFriendRemarkLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateFriendRemarkLogic) UpdateFriendRemark(in *social.UpdateFriendRemarkReq) (*social.UpdateFriendRemarkResp, error) {
	err := l.svcCtx.SocialDao.UpdateFriendRemark(l.ctx, in.Uid, in.PeerUid, in.Remark, in.GroupName)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &social.UpdateFriendRemarkResp{}, nil
}
