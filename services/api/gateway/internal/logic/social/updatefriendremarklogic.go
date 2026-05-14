package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFriendRemarkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFriendRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFriendRemarkLogic {
	return &UpdateFriendRemarkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFriendRemarkLogic) UpdateFriendRemark(req *types.UpdateFriendRemarkReq) (resp *types.UpdateFriendRemarkResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	peerUid, err := strconv.ParseInt(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.UpdateFriendRemark(l.ctx, &socialclient.UpdateFriendRemarkReq{
		Uid:       uid,
		PeerUid:   peerUid,
		Remark:    req.Remark,
		GroupName: req.GroupName,
	})
	if err != nil {
		l.Errorf("UpdateFriendRemark RPC failed: %v", err)
		return nil, err
	}
	return &types.UpdateFriendRemarkResp{Result: true}, nil
}