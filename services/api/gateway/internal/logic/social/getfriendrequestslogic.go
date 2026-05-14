package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendRequestsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFriendRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestsLogic {
	return &GetFriendRequestsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendRequestsLogic) GetFriendRequests(req *types.GetFriendRequestsReq) (resp *types.GetFriendRequestsResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.SocialRpc.GetFriendRequests(l.ctx, &socialclient.GetFriendRequestsReq{
		Uid:  uid,
		Type: req.Type,
		Page: int32(req.Page),
		Size: int32(req.Size),
	})
	if err != nil {
		l.Errorf("GetFriendRequests RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.FriendRequestItem, len(result.List))
	for i, item := range result.List {
		list[i] = types.FriendRequestItem{
			ID:        strconv.FormatInt(item.Id, 10),
			UID:       strconv.FormatInt(item.Uid, 10),
			Name:      item.Name,
			Avatar:    item.Avatar,
			Message:   item.Message,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
		}
	}
	return &types.GetFriendRequestsResp{
		List:  list,
		Total: result.Total,
	}, nil
}