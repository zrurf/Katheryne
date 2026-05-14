package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAnnouncementLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAnnouncementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAnnouncementLogic {
	return &CreateAnnouncementLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAnnouncementLogic) CreateAnnouncement(req *types.CreateAnnouncementReq) (resp *types.CreateAnnouncementResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.CreateAnnouncement(l.ctx, &socialclient.CreateAnnouncementReq{
		GroupId: groupId,
		Uid:     uid,
		Content: req.Content,
	})
	if err != nil {
		l.Errorf("CreateAnnouncement RPC failed: %v", err)
		return nil, err
	}
	return &types.CreateAnnouncementResp{Result: true}, nil
}
