package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFriendRequestLogic {
	return &HandleFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleFriendRequestLogic) HandleFriendRequest(req *types.HandleFriendRequest) (resp *types.HandleFriendResponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	reqId, err := strconv.ParseInt(req.ReqID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.HandleFriendRequest(l.ctx, &socialclient.HandleFriendReq{
		ReqId:      reqId,
		HandlerUid: uid,
		Action:     req.Action,
	})
	if err != nil {
		l.Errorf("HandleFriendRequest RPC failed: %v", err)
		return nil, err
	}
	return &types.HandleFriendResponse{Result: true}, nil
}