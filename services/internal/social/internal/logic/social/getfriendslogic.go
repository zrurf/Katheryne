package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendsLogic {
	return &GetFriendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFriendsLogic) GetFriends(in *social.GetFriendsReq) (*social.GetFriendsResp, error) {
	list, err := l.svcCtx.SocialDao.ListFriends(l.ctx, in.Uid, in.Group)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	items := make([]*social.FriendItem, len(list))
	for i, f := range list {
		user, err := l.svcCtx.UserDBDao.GetUserById(l.ctx, f.PeerUid)
		name := ""
		avatar := ""
		if err == nil && user != nil {
			name = user.Name
			if user.Avatar.Valid {
				avatar = user.Avatar.String
			}
		}
		remark := ""
		if f.Remark.Valid {
			remark = f.Remark.String
		}
		groupName := ""
		if f.GroupName.Valid {
			groupName = f.GroupName.String
		}
		items[i] = &social.FriendItem{
			Uid:       f.PeerUid,
			Name:      name,
			Avatar:    avatar,
			Remark:    remark,
			GroupName: groupName,
		}
	}

	return &social.GetFriendsResp{List: items}, nil
}
