package social

import (
	"context"
	"strconv"
	"time"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type MuteMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMuteMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MuteMemberLogic {
	return &MuteMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MuteMemberLogic) MuteMember(req *types.MuteMemberReq) (resp *types.MuteMemberResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	targetUid, err := strconv.ParseInt(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}
	muteUntil := time.Now().Unix() + int64(req.Duration)
	_, err = l.svcCtx.SocialRpc.MuteMember(l.ctx, &socialclient.MuteMemberReq{
		GroupId:     groupId,
		OperatorUid: uid,
		TargetUid:   targetUid,
		MuteUntil:   muteUntil,
	})
	if err != nil {
		l.Errorf("MuteMember RPC failed: %v", err)
		return nil, err
	}
	return &types.MuteMemberResp{Result: true}, nil
}
