package logic

import (
	"context"

	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsUserBannedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsUserBannedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsUserBannedLogic {
	return &IsUserBannedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsUserBannedLogic) IsUserBanned(in *user.IsUserBannedReq) (*user.IsUserBannedResp, error) {
	if in.Uid <= 0 {
		return &user.IsUserBannedResp{Banned: false}, nil
	}

	// 先查缓存
	banned, reason, err := l.svcCtx.RedisDao.GetBanCache(l.ctx, in.Uid)
	if err == nil {
		return &user.IsUserBannedResp{
			Banned: banned,
			Reason: reason,
		}, nil
	}

	// 查数据库
	ban, err := l.svcCtx.UserDao.GetActiveBan(l.ctx, in.Uid)
	if err != nil {
		l.Errorf("GetActiveBan error: %v", err)
		return nil, err
	}

	isBanned := ban != nil
	var reasonStr string
	if isBanned {
		reasonStr = ban.Reason
	}

	// 写入缓存
	go func() {
		_ = l.svcCtx.RedisDao.SetBanCache(context.Background(), in.Uid, isBanned, reasonStr)
	}()

	return &user.IsUserBannedResp{
		Banned: isBanned,
		Reason: reasonStr,
	}, nil
}
