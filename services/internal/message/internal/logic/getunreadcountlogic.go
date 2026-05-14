package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadCountLogic {
	return &GetUnreadCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUnreadCountLogic) GetUnreadCount(in *message.GetUnreadCountReq) (*message.GetUnreadCountResp, error) {
	cached, err := l.svcCtx.RedisDao.GetUnreadCache(l.ctx, in.Uid, in.ConvId)
	if err != nil {
		l.Logger.Error("get unread cache error:", err)
	}
	if cached >= 0 {
		return &message.GetUnreadCountResp{Count: cached}, nil
	}

	count, err := l.svcCtx.MessageDao.GetUnreadCount(l.ctx, in.ConvId, in.Uid)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.SetUnreadCache(l.ctx, in.Uid, in.ConvId, count)
	if err != nil {
		l.Logger.Error("set unread cache error:", err)
	}

	return &message.GetUnreadCountResp{Count: count}, nil
}
