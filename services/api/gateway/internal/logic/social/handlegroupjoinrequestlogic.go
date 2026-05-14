package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupJoinRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleGroupJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupJoinRequestLogic {
	return &HandleGroupJoinRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGroupJoinRequestLogic) HandleGroupJoinRequest(req *types.HandleGroupJoinReq) (resp *types.HandleGroupJoinResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	reqId, err := strconv.ParseInt(req.ReqID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.HandleGroupJoinRequest(l.ctx, &socialclient.HandleGroupJoinReq{
		ReqId:       reqId,
		ReviewerUid: uid,
		Action:      req.Action,
	})
	if err != nil {
		l.Errorf("HandleGroupJoinRequest RPC failed: %v", err)
		return nil, err
	}
	return &types.HandleGroupJoinResp{Result: true}, nil
}
