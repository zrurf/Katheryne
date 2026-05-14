package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupJoinRequestsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupJoinRequestsLogic {
	return &GetGroupJoinRequestsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupJoinRequestsLogic) GetGroupJoinRequests(req *types.GetGroupJoinRequestsReq) (resp *types.GetGroupJoinRequestsResp, err error) {
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.SocialRpc.GetGroupJoinRequests(l.ctx, &socialclient.GetGroupJoinRequestsReq{
		GroupId: groupId,
		Status:  req.Status,
	})
	if err != nil {
		l.Errorf("GetGroupJoinRequests RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.GroupJoinRequestItem, len(result.List))
	for i, item := range result.List {
		list[i] = types.GroupJoinRequestItem{
			ID:        strconv.FormatInt(item.Id, 10),
			UID:       strconv.FormatInt(item.Uid, 10),
			Name:      item.Name,
			Avatar:    item.Avatar,
			Message:   item.Message,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
		}
	}
	return &types.GetGroupJoinRequestsResp{
		List: list,
	}, nil
}
