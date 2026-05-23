package auth

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileReq) (resp *types.UpdateProfileResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	_, err = l.svcCtx.UserRpc.UpdateUser(l.ctx, &userclient.UpdateUserReq{
		Uid:    uid,
		Name:   req.Name,
		Avatar: req.Avatar,
	})
	if err != nil {
		l.Errorf("UpdateProfile RPC failed: %v", err)
		return nil, err
	}
	return &types.UpdateProfileResp{
		Name:   req.Name,
		Avatar: req.Avatar,
	}, nil
}
