package logic

import (
	"context"
	"errors"

	"social/internal/svc"
	"social/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupJoinRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleGroupJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupJoinRequestLogic {
	return &HandleGroupJoinRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HandleGroupJoinRequestLogic) HandleGroupJoinRequest(in *social.HandleGroupJoinReq) (*social.HandleGroupJoinResp, error) {
	req, err := l.svcCtx.SocialDao.GetGroupJoinRequestById(l.ctx, in.ReqId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if req.Status != "pending" {
		return nil, errors.New("该申请已处理")
	}

	operator, err := l.svcCtx.SocialDao.GetGroupMember(l.ctx, req.GroupId, in.ReviewerUid)
	if err != nil {
		l.Logger.Error(err)
		return nil, errors.New("无权处理该申请")
	}

	if operator.Role != "OWNER" && operator.Role != "ADMIN" {
		return nil, errors.New("无权处理该申请")
	}

	err = l.svcCtx.SocialDao.UpdateGroupJoinRequestStatus(l.ctx, in.ReqId, in.Action)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if in.Action == "accept" {
		err = l.svcCtx.SocialDao.AddGroupMember(l.ctx, req.GroupId, req.Uid, "MEMBER", "", 0)
		if err != nil {
			l.Logger.Error(err)
			return nil, err
		}

		err = l.svcCtx.SocialDao.IncrGroupMemberCount(l.ctx, req.GroupId, 1)
		if err != nil {
			l.Logger.Error(err)
		}

		convId, err := l.svcCtx.SocialDao.GetConversationByGroupId(l.ctx, req.GroupId)
		if err == nil {
			err = l.svcCtx.SocialDao.AddConvMember(l.ctx, convId, req.Uid)
			if err != nil {
				l.Logger.Error(err)
			}
		}

		err = l.svcCtx.RedisDao.DelGroupMembersCache(l.ctx, req.GroupId)
		if err != nil {
			l.Logger.Error("del group members cache error:", err)
		}
	}

	return &social.HandleGroupJoinResp{}, nil
}
