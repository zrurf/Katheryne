package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnnouncementsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAnnouncementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnnouncementsLogic {
	return &GetAnnouncementsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAnnouncementsLogic) GetAnnouncements(req *types.GetAnnouncementsReq) (resp *types.GetAnnouncementsResp, err error) {
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.SocialRpc.GetAnnouncements(l.ctx, &socialclient.GetAnnouncementsReq{
		GroupId: groupId,
		Page:    int32(req.Page),
		Size:    int32(req.Size),
	})
	if err != nil {
		l.Errorf("GetAnnouncements RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.AnnouncementItem, len(result.List))
	for i, item := range result.List {
		list[i] = types.AnnouncementItem{
			ID:        strconv.FormatInt(item.Id, 10),
			UID:       strconv.FormatInt(item.Uid, 10),
			Name:      item.Name,
			Content:   item.Content,
			Pinned:    item.Pinned,
			CreatedAt: item.CreatedAt,
		}
	}
	return &types.GetAnnouncementsResp{
		List:  list,
		Total: result.Total,
	}, nil
}
