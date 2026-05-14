package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveBlacklistLogic {
	return &RemoveBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveBlacklistLogic) RemoveBlacklist(req *types.RemoveBlacklistReq) (resp *types.RemoveBlacklistResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	peerUid, err := strconv.ParseInt(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.RemoveBlacklist(l.ctx, &socialclient.RemoveBlacklistReq{
		Uid:     uid,
		PeerUid: peerUid,
	})
	if err != nil {
		l.Errorf("RemoveBlacklist RPC failed: %v", err)
		return nil, err
	}
	return &types.RemoveBlacklistResp{Result: true}, nil
}