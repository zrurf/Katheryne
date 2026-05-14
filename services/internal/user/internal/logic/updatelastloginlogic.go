package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLastLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLastLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLastLoginLogic {
	return &UpdateLastLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateLastLoginLogic) UpdateLastLogin(in *user.UpdateLastLoginReq) (*user.UpdateLastLoginResp, error) {
	if in.Uid <= 0 {
		return &user.UpdateLastLoginResp{}, nil
	}

	err := l.svcCtx.UserDao.UpdateLastLogin(l.ctx, in.Uid)
	if err != nil {
		l.Errorf("UpdateLastLogin error: %v", err)
		return nil, err
	}

	// 删除缓存，让下次读取时刷新 last_login
	go func() {
		_ = l.svcCtx.RedisDao.DelUserCache(context.Background(), in.Uid)
	}()

	return &user.UpdateLastLoginResp{}, nil
}
