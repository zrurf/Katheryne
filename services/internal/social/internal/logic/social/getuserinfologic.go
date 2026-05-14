package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *social.GetUserInfoReq) (*social.GetUserInfoResp, error) {
	user, err := l.svcCtx.UserDBDao.GetUserById(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	avatar := ""
	if user.Avatar.Valid {
		avatar = user.Avatar.String
	}
	return &social.GetUserInfoResp{
		Info: &social.UserInfo{
			Uid:      user.Uid,
			Nickname: user.Name,
			Avatar:   avatar,
		},
	}, nil
}
