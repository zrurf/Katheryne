package logic

import (
	"context"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetGroupNickLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetGroupNickLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetGroupNickLogic {
	return &SetGroupNickLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetGroupNickLogic) SetGroupNick(in *social.SetGroupNickReq) (*social.SetGroupNickResp, error) {
	err := l.svcCtx.SocialDao.UpdateMemberNick(l.ctx, in.GroupId, in.Uid, in.Nick)
	if err != nil {
		return nil, err
	}
	return &social.SetGroupNickResp{}, nil
}
