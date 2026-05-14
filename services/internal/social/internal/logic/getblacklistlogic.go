package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlacklistLogic {
	return &GetBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBlacklistLogic) GetBlacklist(in *social.GetBlacklistReq) (*social.GetBlacklistResp, error) {
	list, err := l.svcCtx.SocialDao.GetBlacklist(l.ctx, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	items := make([]*social.FriendItem, len(list))
	for i, b := range list {
		items[i] = &social.FriendItem{
			Uid: b.PeerUid,
		}
	}

	return &social.GetBlacklistResp{List: items}, nil
}
