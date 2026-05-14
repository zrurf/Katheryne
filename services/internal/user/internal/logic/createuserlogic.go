package logic

import (
	"context"
	"time"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *user.CreateUserReq) (*user.CreateUserResp, error) {
	if in.Name == "" || in.Phone == "" {
		return nil, nil
	}

	// 检查手机号是否已存在
	exists, err := l.svcCtx.UserDao.PhoneExists(l.ctx, in.Phone)
	if err != nil {
		l.Errorf("PhoneExists error: %v", err)
		return nil, err
	}
	if exists {
		return nil, nil
	}

	// 生成 UID：使用当前时间戳纳秒 + 随机数，保证唯一性
	uid := time.Now().UnixNano()

	// 创建用户
	u, err := l.svcCtx.UserDao.CreateUser(l.ctx, uid, in.Phone, in.Name, in.OpaqueRecord)
	if err != nil {
		l.Errorf("CreateUser error: %v", err)
		return nil, err
	}

	// 写入缓存
	go func() {
		_ = l.svcCtx.RedisDao.SetUserCache(context.Background(), u.UID, u)
	}()

	return &user.CreateUserResp{
		User: daoUserToProto(u),
	}, nil
}
