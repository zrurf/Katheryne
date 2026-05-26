package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBlacklistLogic {
	return &AddBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddBlacklistLogic) AddBlacklist(req *types.AddBlacklistReq) (resp *types.AddBlacklistResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	peerUid, err := strconv.ParseInt(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.AddBlacklist(l.ctx, &socialclient.AddBlacklistReq{
		Uid:     uid,
		PeerUid: peerUid,
	})
	if err != nil {
		l.Errorf("AddBlacklist RPC failed: %v", err)
		return nil, err
	}
	return &types.AddBlacklistResp{Result: true}, nil
}
