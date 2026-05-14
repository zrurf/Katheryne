package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type LeaveGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLeaveGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveGroupLogic {
	return &LeaveGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LeaveGroupLogic) LeaveGroup(req *types.LeaveGroupReq) (resp *types.LeaveGroupResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.LeaveGroup(l.ctx, &socialclient.LeaveGroupReq{
		GroupId: groupId,
		Uid:     uid,
	})
	if err != nil {
		l.Errorf("LeaveGroup RPC failed: %v", err)
		return nil, err
	}
	return &types.LeaveGroupResp{Result: true}, nil
}