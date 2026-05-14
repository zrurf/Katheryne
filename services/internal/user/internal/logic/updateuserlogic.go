package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserReq) (*user.UpdateUserResp, error) {
	if in.Uid <= 0 {
		return &user.UpdateUserResp{}, nil
	}

	err := l.svcCtx.UserDao.UpdateUser(l.ctx, in.Uid, in.Name, in.Avatar)
	if err != nil {
		l.Errorf("UpdateUser error: %v", err)
		return nil, err
	}

	// 删除缓存，下次读取时重新加载
	go func() {
		_ = l.svcCtx.RedisDao.DelUserCache(context.Background(), in.Uid)
	}()

	return &user.UpdateUserResp{}, nil
}
