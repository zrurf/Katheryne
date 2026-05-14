package logic

import (
	"context"

	"user/internal/dao"
	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByUIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByUIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByUIDLogic {
	return &GetUserByUIDLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByUIDLogic) GetUserByUID(in *user.GetUserByUIDReq) (*user.GetUserResp, error) {
	if in.Uid <= 0 {
		return &user.GetUserResp{}, nil
	}

	// 先查缓存
	u, err := l.svcCtx.RedisDao.GetUserCache(l.ctx, in.Uid)
	if err == nil && u != nil {
		return &user.GetUserResp{
			User: daoUserToProto(u),
		}, nil
	}

	// 再查数据库
	u, err = l.svcCtx.UserDao.GetUserByUID(l.ctx, in.Uid)
	if err != nil {
		l.Errorf("GetUserByUID db error: %v", err)
		return nil, err
	}
	if u == nil {
		return &user.GetUserResp{}, nil
	}

	// 写入缓存（异步，不阻塞返回）
	go func() {
		_ = l.svcCtx.RedisDao.SetUserCache(context.Background(), in.Uid, u)
	}()

	return &user.GetUserResp{
		User: daoUserToProto(u),
	}, nil
}

func daoUserToProto(u *dao.User) *user.UserInfo {
	if u == nil {
		return nil
	}
	var avatar string
	if u.Avatar != nil {
		avatar = *u.Avatar
	}
	return &user.UserInfo{
		Uid:       u.UID,
		Name:      u.Name,
		Phone:     u.Phone,
		Avatar:    avatar,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
		LastLogin: u.LastLogin.Unix(),
	}
}
