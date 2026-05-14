package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type SetConvMuteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetConvMuteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConvMuteLogic {
	return &SetConvMuteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetConvMuteLogic) SetConvMute(in *conversation.SetConvMuteReq) (*conversation.SetConvMuteResp, error) {
	member, err := l.svcCtx.ConversationDao.GetConvMember(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not a member of this conversation")
		}
		l.Logger.Error(err)
		return nil, err
	}
	if !member.IsActive {
		return nil, errors.New("not an active member of this conversation")
	}

	err = l.svcCtx.ConversationDao.SetConvMemberMute(l.ctx, in.ConvId, in.Uid, in.Mute)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, in.Uid)
	_ = l.svcCtx.RedisDao.DelConvCache(l.ctx, in.ConvId)

	return &conversation.SetConvMuteResp{}, nil
}
