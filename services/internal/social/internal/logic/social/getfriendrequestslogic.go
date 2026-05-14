package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestsLogic {
	return &GetFriendRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFriendRequestsLogic) GetFriendRequests(in *social.GetFriendRequestsReq) (*social.GetFriendRequestsResp, error) {
	list, total, err := l.svcCtx.SocialDao.ListFriendRequests(l.ctx, in.Uid, in.Type, in.Page, in.Size)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	items := make([]*social.FriendRequestItem, len(list))
	for i, r := range list {
		msg := ""
		if r.Message.Valid {
			msg = r.Message.String
		}
		name := ""
		avatar := ""
		user, err := l.svcCtx.UserDBDao.GetUserById(l.ctx, r.Uid)
		if err == nil && user != nil {
			name = user.Name
			avatar = nullString(user.Avatar)
		}
		items[i] = &social.FriendRequestItem{
			Id:        r.Id,
			Uid:       r.Uid,
			Name:      name,
			Avatar:    avatar,
			Message:   msg,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.UnixMilli(),
		}
	}

	return &social.GetFriendRequestsResp{
		List:  items,
		Total: total,
	}, nil
}
