package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGroupLogic) CreateGroup(req *types.CreateGroupReq) (resp *types.CreateGroupResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	memberUids := make([]int64, len(req.MemberUIDs))
	for i, s := range req.MemberUIDs {
		memberUids[i], err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	result, err := l.svcCtx.SocialRpc.CreateGroup(l.ctx, &socialclient.CreateGroupReq{
		OwnerUid:   uid,
		Name:       req.Name,
		Avatar:     req.Avatar,
		MemberUids: memberUids,
		VerifyMode: req.VerifyMode,
	})
	if err != nil {
		l.Errorf("CreateGroup RPC failed: %v", err)
		return nil, err
	}
	return &types.CreateGroupResp{
		GroupID: strconv.FormatInt(result.GroupId, 10),
		ConvID:  strconv.FormatInt(result.ConvId, 10),
	}, nil
}