package sociallogic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupJoinRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupJoinRequestsLogic {
	return &GetGroupJoinRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupJoinRequestsLogic) GetGroupJoinRequests(in *social.GetGroupJoinRequestsReq) (*social.GetGroupJoinRequestsResp, error) {
	requests, total, err := l.svcCtx.SocialDao.ListGroupJoinRequests(l.ctx, in.GroupId, in.Status, in.Page, in.Size)
	if err != nil {
		return nil, err
	}

	list := make([]*social.GroupJoinRequestItem, len(requests))
	for i, r := range requests {
		msg := ""
		if r.Message.Valid {
			msg = r.Message.String
		}
		list[i] = &social.GroupJoinRequestItem{
			Id:        r.Id,
			Uid:       r.Uid,
			Message:   msg,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Unix(),
		}
	}
	return &social.GetGroupJoinRequestsResp{
		List:  list,
		Total: total,
	}, nil
}