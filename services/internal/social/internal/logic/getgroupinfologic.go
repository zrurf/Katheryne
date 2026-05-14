package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoLogic {
	return &GetGroupInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupInfoLogic) GetGroupInfo(in *social.GetGroupInfoReq) (*social.GetGroupInfoResp, error) {
	g, err := l.svcCtx.SocialDao.GetGroupById(l.ctx, in.GroupId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return &social.GetGroupInfoResp{
		Info: &social.GroupInfo{
			GroupId:     g.GroupId,
			Name:        g.Name,
			Avatar:      nullString(g.Avatar),
			Owner:       g.Owner,
			MemberCount: g.MemberCount,
			Status:      g.Status,
			VerifyMode:  g.VerifyMode,
			CreatedAt:   g.CreatedAt.UnixMilli(),
		},
	}, nil
}
