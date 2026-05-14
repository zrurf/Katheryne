package logic

import (
	"context"

	"user/internal/dao"
	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserInfoLogic {
	return &BatchGetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchGetUserInfoLogic) BatchGetUserInfo(in *user.BatchGetUserInfoReq) (*user.BatchGetUserInfoResp, error) {
	if len(in.Uids) == 0 {
		return &user.BatchGetUserInfoResp{Users: make(map[int64]*user.UserInfo)}, nil
	}

	// 限制批量数量
	uids := in.Uids
	if len(uids) > 100 {
		uids = uids[:100]
	}

	result := make(map[int64]*user.UserInfo, len(uids))
	missUids := make([]int64, 0, len(uids))

	// 先查缓存
	for _, uid := range uids {
		u, err := l.svcCtx.RedisDao.GetUserCache(l.ctx, uid)
		if err == nil && u != nil {
			result[uid] = daoUserToProto(u)
		} else {
			missUids = append(missUids, uid)
		}
	}

	// 缓存未命中查数据库
	if len(missUids) > 0 {
		users, err := l.svcCtx.UserDao.GetUsersByUIDs(l.ctx, missUids)
		if err != nil {
			l.Errorf("BatchGetUserInfo db error: %v", err)
			return nil, err
		}
		for _, u := range users {
			result[u.UID] = daoUserToProto(u)
			// 异步写入缓存
			go func(u *dao.User) {
				_ = l.svcCtx.RedisDao.SetUserCache(context.Background(), u.UID, u)
			}(u)
		}
	}

	return &user.BatchGetUserInfoResp{Users: result}, nil
}
