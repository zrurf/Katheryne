package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnnouncementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAnnouncementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnnouncementsLogic {
	return &GetAnnouncementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAnnouncementsLogic) GetAnnouncements(in *social.GetAnnouncementsReq) (*social.GetAnnouncementsResp, error) {
	announcements, total, err := l.svcCtx.SocialDao.ListAnnouncements(l.ctx, in.GroupId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}

	list := make([]*social.AnnouncementItem, len(announcements))
	for i, a := range announcements {
		name := ""
		user, err := l.svcCtx.UserDBDao.GetUserById(l.ctx, a.Uid)
		if err == nil && user != nil {
			name = user.Name
		}
		list[i] = &social.AnnouncementItem{
			Id:        a.Id,
			Uid:       a.Uid,
			Name:      name,
			Content:   a.Content,
			Pinned:    a.Pinned,
			CreatedAt: a.CreatedAt.UnixMilli(),
		}
	}
	return &social.GetAnnouncementsResp{
		List:  list,
		Total: total,
	}, nil
}
