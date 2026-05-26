package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoLogic {
	return &GetGroupInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupInfoLogic) GetGroupInfo(req *types.GetGroupInfoReq) (resp *types.GroupInfoResp, err error) {
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.SocialRpc.GetGroupInfo(l.ctx, &socialclient.GetGroupInfoReq{
		GroupId: groupId,
	})
	if err != nil {
		l.Errorf("GetGroupInfo RPC failed: %v", err)
		return nil, err
	}
	return &types.GroupInfoResp{
		GroupID:     strconv.FormatInt(result.Info.GroupId, 10),
		Name:        result.Info.Name,
		Avatar:      result.Info.Avatar,
		Owner:       strconv.FormatInt(result.Info.Owner, 10),
		MemberCount: int(result.Info.MemberCount),
		Status:      result.Info.Status,
		VerifyMode:  result.Info.VerifyMode,
		CreatedAt:   result.Info.CreatedAt,
	}, nil
}
