package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupInviteLogic {
	return &HandleGroupInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGroupInviteLogic) HandleGroupInvite(req *types.HandleGroupInviteReq) (resp *types.HandleGroupInviteResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	inviteId, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.HandleGroupInvite(l.ctx, &socialclient.HandleGroupInviteReq{
		InviteId: inviteId,
		Uid:      uid,
		Action:   req.Action,
	})
	if err != nil {
		l.Errorf("HandleGroupInvite RPC failed: %v", err)
		return nil, err
	}
	return &types.HandleGroupInviteResp{Result: true}, nil
}
