package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendFriendRequestLogic {
	return &SendFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendFriendRequestLogic) SendFriendRequest(req *types.SendFriendRequest) (resp *types.SendFriendResponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	toUid, err := strconv.ParseInt(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.SendFriendRequest(l.ctx, &socialclient.SendFriendReq{
		Uid:     uid,
		ToUid:   toUid,
		Message: req.Message,
	})
	if err != nil {
		l.Errorf("SendFriendRequest RPC failed: %v", err)
		return nil, err
	}
	return &types.SendFriendResponse{Result: true}, nil
}
