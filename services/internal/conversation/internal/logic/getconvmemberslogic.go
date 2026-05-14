package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConvMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvMembersLogic {
	return &GetConvMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConvMembersLogic) GetConvMembers(in *conversation.GetConvMembersReq) (*conversation.GetConvMembersResp, error) {
	uids, err := l.svcCtx.RedisDao.GetConvMembersCache(l.ctx, in.ConvId)
	if err == nil && len(uids) > 0 {
		return &conversation.GetConvMembersResp{Uids: uids}, nil
	}

	uids, err = l.svcCtx.ConversationDao.ListConvMembers(l.ctx, in.ConvId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	_ = l.svcCtx.RedisDao.SetConvMembersCache(l.ctx, in.ConvId, uids)

	return &conversation.GetConvMembersResp{Uids: uids}, nil
}
