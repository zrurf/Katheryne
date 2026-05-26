package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendsLogic {
	return &GetFriendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendsLogic) GetFriends(req *types.GetFriendsReq) (resp *types.GetFriendsResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.SocialRpc.GetFriends(l.ctx, &socialclient.GetFriendsReq{
		Uid:   uid,
		Group: req.Group,
	})
	if err != nil {
		l.Errorf("GetFriends RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.FriendItem, len(result.List))
	for i, item := range result.List {
		list[i] = types.FriendItem{
			UID:       strconv.FormatInt(item.Uid, 10),
			Name:      item.Name,
			Avatar:    item.Avatar,
			Remark:    item.Remark,
			Status:    item.Status,
			GroupName: item.GroupName,
		}
	}
	return &types.GetFriendsResp{List: list}, nil
}
