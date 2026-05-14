package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupLogic {
	return &UpdateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateGroupLogic) UpdateGroup(req *types.UpdateGroupReq) (resp *types.UpdateGroupResp, err error) {
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.UpdateGroup(l.ctx, &socialclient.UpdateGroupReq{
		GroupId:    groupId,
		Name:       req.Name,
		Avatar:     req.Avatar,
		VerifyMode: req.VerifyMode,
	})
	if err != nil {
		l.Errorf("UpdateGroup RPC failed: %v", err)
		return nil, err
	}
	return &types.UpdateGroupResp{Result: true}, nil
}