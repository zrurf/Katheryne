package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGroupNickLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateGroupNickLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupNickLogic {
	return &UpdateGroupNickLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateGroupNickLogic) UpdateGroupNick(req *types.UpdateGroupNickReq) (resp *types.UpdateGroupNickResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.SetGroupNick(l.ctx, &socialclient.SetGroupNickReq{
		GroupId: groupId,
		Uid:     uid,
		Nick:    req.Nick,
	})
	if err != nil {
		l.Errorf("SetGroupNick RPC failed: %v", err)
		return nil, err
	}
	return &types.UpdateGroupNickResp{Result: true}, nil
}
