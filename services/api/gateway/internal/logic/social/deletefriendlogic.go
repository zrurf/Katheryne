package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFriendLogic) DeleteFriend(req *types.DeleteFriendReq) (resp *types.DeleteFriendResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	peerUid, err := strconv.ParseInt(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.DeleteFriend(l.ctx, &socialclient.DeleteFriendReq{
		Uid:     uid,
		PeerUid: peerUid,
	})
	if err != nil {
		l.Errorf("DeleteFriend RPC failed: %v", err)
		return nil, err
	}
	return &types.DeleteFriendResp{Result: true}, nil
}