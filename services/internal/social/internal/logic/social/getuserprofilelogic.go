package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileLogic {
	return &GetUserProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserProfileLogic) GetUserProfile(in *social.GetUserProfileReq) (*social.GetUserProfileResp, error) {
	user, err := l.svcCtx.UserDBDao.GetUserById(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	avatar := ""
	if user.Avatar.Valid {
		avatar = user.Avatar.String
	}
	return &social.GetUserProfileResp{
		Profile: &social.UserInfo{
			Uid:      user.Uid,
			Nickname: user.Name,
			Avatar:   avatar,
		},
	}, nil
}
