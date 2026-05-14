package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMembersLogic {
	return &GetGroupMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupMembersLogic) GetGroupMembers(req *types.GetGroupMembersReq) (resp *types.GetGroupMembersResp, err error) {
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.SocialRpc.GetGroupMembers(l.ctx, &socialclient.GetGroupMembersReq{
		GroupId: groupId,
		Role:    req.Role,
	})
	if err != nil {
		l.Errorf("GetGroupMembers RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.GroupMemberItem, len(result.List))
	for i, item := range result.List {
		list[i] = types.GroupMemberItem{
			UID:       strconv.FormatInt(item.Uid, 10),
			Name:      item.Name,
			Avatar:    item.Avatar,
			Role:      item.Role,
			Nick:      item.Nick,
			JoinTime:  item.JoinTime,
			MuteUntil: item.MuteUntil,
		}
	}
	return &types.GetGroupMembersResp{
		List: list,
	}, nil
}
