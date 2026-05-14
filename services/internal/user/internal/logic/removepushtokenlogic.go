package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemovePushTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemovePushTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePushTokenLogic {
	return &RemovePushTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemovePushTokenLogic) RemovePushToken(in *user.RemovePushTokenReq) (*user.RemovePushTokenResp, error) {
	if in.Uid <= 0 || in.DeviceId == "" {
		return &user.RemovePushTokenResp{}, nil
	}

	err := l.svcCtx.UserDao.RemoveUserDevice(l.ctx, in.Uid, in.DeviceId)
	if err != nil {
		l.Errorf("RemoveUserDevice error: %v", err)
		return nil, err
	}

	// 删除相关缓存
	go func() {
		_ = l.svcCtx.RedisDao.DelDevicesCache(context.Background(), in.Uid)
	}()

	return &user.RemovePushTokenResp{}, nil
}
