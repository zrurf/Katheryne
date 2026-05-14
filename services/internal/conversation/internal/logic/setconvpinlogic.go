package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type SetConvPinLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetConvPinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConvPinLogic {
	return &SetConvPinLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetConvPinLogic) SetConvPin(in *conversation.SetConvPinReq) (*conversation.SetConvPinResp, error) {
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

	err = l.svcCtx.ConversationDao.SetConvMemberPin(l.ctx, in.ConvId, in.Uid, in.Pinned)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, in.Uid)
	_ = l.svcCtx.RedisDao.DelConvCache(l.ctx, in.ConvId)

	return &conversation.SetConvPinResp{}, nil
}
