package sociallogic

import (
	"context"
	"errors"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFriendRequestLogic {
	return &HandleFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HandleFriendRequestLogic) HandleFriendRequest(in *social.HandleFriendReq) (*social.HandleFriendResp, error) {
	req, err := l.svcCtx.SocialDao.GetFriendRequestById(l.ctx, in.ReqId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if req.ToUid != in.HandlerUid {
		return nil, errors.New("无权处理该好友请求")
	}

	if req.Status != "pending" {
		return nil, errors.New("该请求已处理")
	}

	status := in.Action
	if status == "accept" {
		status = "accepted"
	} else if status == "reject" {
		status = "rejected"
	}

	err = l.svcCtx.SocialDao.UpdateFriendRequestStatus(l.ctx, in.ReqId, status)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if in.Action == "accept" {
		err = l.svcCtx.SocialDao.AddFriendship(l.ctx, req.Uid, req.ToUid, "", "")
		if err != nil {
			l.Logger.Error(err)
			return nil, err
		}
	}

	return &social.HandleFriendResp{}, nil
}
