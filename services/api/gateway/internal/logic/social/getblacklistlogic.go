package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlacklistLogic {
	return &GetBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlacklistLogic) GetBlacklist() (resp *types.GetBlacklistResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.SocialRpc.GetBlacklist(l.ctx, &socialclient.GetBlacklistReq{
		Uid: uid,
	})
	if err != nil {
		l.Errorf("GetBlacklist RPC failed: %v", err)
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
	return &types.GetBlacklistResp{List: list}, nil
}
