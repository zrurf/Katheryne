package logic

import (
	"context"
	"errors"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrCreateSingleConvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrCreateSingleConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateSingleConvLogic {
	return &GetOrCreateSingleConvLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrCreateSingleConvLogic) GetOrCreateSingleConv(in *conversation.GetOrCreateSingleConvReq) (*conversation.GetOrCreateSingleConvResp, error) {
	if in.Uid == in.PeerUid {
		return nil, errors.New("cannot create conversation with self")
	}

	uid1, uid2 := in.Uid, in.PeerUid
	if uid1 > uid2 {
		uid1, uid2 = uid2, uid1
	}

	conv, err := l.svcCtx.ConversationDao.GetSingleConversationByPair(l.ctx, uid1, uid2)
	if err == nil {
		_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, in.Uid)
		_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, in.PeerUid)
		return &conversation.GetOrCreateSingleConvResp{
			ConvId:  conv.ConvId,
			Created: false,
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		l.Logger.Error(err)
		return nil, err
	}

	convId, err := l.svcCtx.ConversationDao.CreateSingleConversation(l.ctx, uid1, uid2)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.ConversationDao.UpsertConvMember(l.ctx, convId, uid1, false, false, true)
	if err != nil {
		l.Logger.Error(err)
	}
	err = l.svcCtx.ConversationDao.UpsertConvMember(l.ctx, convId, uid2, false, false, true)
	if err != nil {
		l.Logger.Error(err)
	}

	_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, in.Uid)
	_ = l.svcCtx.RedisDao.DelConvListCache(l.ctx, in.PeerUid)

	return &conversation.GetOrCreateSingleConvResp{
		ConvId:  convId,
		Created: true,
	}, nil
}
