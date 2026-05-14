package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type InviteToGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInviteToGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteToGroupLogic {
	return &InviteToGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InviteToGroupLogic) InviteToGroup(req *types.InviteToGroupReq) (resp *types.InviteToGroupResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	inviteeUids := make([]int64, len(req.InviteeUIDs))
	for i, s := range req.InviteeUIDs {
		inviteeUids[i], err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	result, err := l.svcCtx.SocialRpc.InviteToGroup(l.ctx, &socialclient.InviteToGroupReq{
		GroupId:     groupId,
		InviterUid:  uid,
		InviteeUids: inviteeUids,
		Message:     req.Message,
	})
	if err != nil {
		l.Errorf("InviteToGroup RPC failed: %v", err)
		return nil, err
	}

	failedUIDs := make([]string, len(result.FailedUids))
	for i, uid := range result.FailedUids {
		failedUIDs[i] = strconv.FormatInt(uid, 10)
	}
	return &types.InviteToGroupResp{FailedUIDs: failedUIDs}, nil
}
