package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserStatusLogic {
	return &UpdateUserStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserStatusLogic) UpdateUserStatus(in *user.UpdateUserStatusReq) (*user.UpdateUserStatusResp, error) {
	if in.Uid <= 0 || in.Status == "" {
		return &user.UpdateUserStatusResp{}, nil
	}

	err := l.svcCtx.UserDao.UpdateUserStatus(l.ctx, in.Uid, in.Status)
	if err != nil {
		l.Errorf("UpdateUserStatus error: %v", err)
		return nil, err
	}

	// 删除用户缓存和封禁缓存
	go func() {
		_ = l.svcCtx.RedisDao.DelUserCache(context.Background(), in.Uid)
		_ = l.svcCtx.RedisDao.DelBanCache(context.Background(), in.Uid)
	}()

	return &user.UpdateUserStatusResp{}, nil
}
