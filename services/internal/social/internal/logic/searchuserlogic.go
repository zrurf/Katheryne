package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUserLogic {
	return &SearchUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchUserLogic) SearchUser(in *social.SearchUserReq) (*social.SearchUserResp, error) {
	users, total, err := l.svcCtx.UserDBDao.SearchUser(l.ctx, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}

	list := make([]*social.UserInfo, len(users))
	for i, u := range users {
		list[i] = &social.UserInfo{
			Uid:      u.Uid,
			Nickname: u.Name,
			Avatar:   u.Avatar.String,
		}
	}
	return &social.SearchUserResp{
		List:  list,
		Total: total,
	}, nil
}
