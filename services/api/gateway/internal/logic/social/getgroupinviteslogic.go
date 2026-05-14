package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInvitesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupInvitesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInvitesLogic {
	return &GetGroupInvitesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupInvitesLogic) GetGroupInvites(req *types.GetGroupInvitesReq) (resp *types.GetGroupInvitesResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.SocialRpc.GetGroupInvites(l.ctx, &socialclient.GetGroupInvitesReq{
		Uid: uid,
	})
	if err != nil {
		l.Errorf("GetGroupInvites RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.GroupInviteItem, len(result.List))
	for i, item := range result.List {
		list[i] = types.GroupInviteItem{
			ID:          strconv.FormatInt(item.Id, 10),
			GroupID:     strconv.FormatInt(item.GroupId, 10),
			GroupName:   item.GroupName,
			GroupAvatar: item.GroupAvatar,
			InviterUID:  strconv.FormatInt(item.InviterUid, 10),
			InviterName: item.InviterName,
			Message:     item.Message,
			CreatedAt:   item.CreatedAt,
		}
	}
	return &types.GetGroupInvitesResp{
		List: list,
	}, nil
}
