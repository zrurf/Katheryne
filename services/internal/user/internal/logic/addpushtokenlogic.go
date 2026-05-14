package logic

import (
	"context"

	"user/internal/dao"
	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddPushTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddPushTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPushTokenLogic {
	return &AddPushTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddPushTokenLogic) AddPushToken(in *user.AddPushTokenReq) (*user.AddPushTokenResp, error) {
	if in.Uid <= 0 || in.DeviceId == "" || in.Token == "" {
		return &user.AddPushTokenResp{}, nil
	}

	// 写入设备表
	dev := &dao.UserDevice{
		UID:       in.Uid,
		DeviceID:  in.DeviceId,
		PushToken: &in.Token,
	}
	if in.Platform != "" {
		dev.Platform = &in.Platform
	}

	err := l.svcCtx.UserDao.UpsertUserDevice(l.ctx, dev)
	if err != nil {
		l.Errorf("UpsertUserDevice error: %v", err)
		return nil, err
	}

	// 同时写入 Redis 集合，方便快速获取推送令牌
	go func() {
		_ = l.svcCtx.RedisDao.AddPushToken(context.Background(), in.Uid, in.Token)
		_ = l.svcCtx.RedisDao.DelDevicesCache(context.Background(), in.Uid)
	}()

	return &user.AddPushTokenResp{}, nil
}
