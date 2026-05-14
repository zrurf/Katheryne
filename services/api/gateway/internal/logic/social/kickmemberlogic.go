package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type KickMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKickMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KickMemberLogic {
	return &KickMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KickMemberLogic) KickMember(req *types.KickMemberReq) (resp *types.KickMemberResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	targetUid, err := strconv.ParseInt(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.KickMember(l.ctx, &socialclient.KickMemberReq{
		GroupId:     groupId,
		OperatorUid: uid,
		TargetUid:   targetUid,
	})
	if err != nil {
		l.Errorf("KickMember RPC failed: %v", err)
		return nil, err
	}
	return &types.KickMemberResp{Result: true}, nil
}
