package logic

import (
	"context"

	"user/internal/dao"
	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserSettingsLogic {
	return &UpdateUserSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserSettingsLogic) UpdateUserSettings(in *user.UpdateUserSettingsReq) (*user.UpdateUserSettingsResp, error) {
	if in.Uid <= 0 {
		return &user.UpdateUserSettingsResp{}, nil
	}

	cfg := &dao.UserConfig{
		UID:             in.Uid,
		Language:        in.Language,
		MsgNotification: in.MsgNotification,
		SoundEnabled:    in.SoundEnabled,
		AutoTranslate:   in.AutoTranslate,
		ContentFilter:   in.ContentFilter,
	}
	if in.TranslateTarget != "" {
		target := in.TranslateTarget
		cfg.TranslateTarget = &target
	}

	err := l.svcCtx.UserDao.UpsertUserConfig(l.ctx, cfg)
	if err != nil {
		l.Errorf("UpsertUserConfig error: %v", err)
		return nil, err
	}

	// 删除配置缓存
	go func() {
		_ = l.svcCtx.RedisDao.DelUserSettingsCache(context.Background(), in.Uid)
	}()

	return &user.UpdateUserSettingsResp{}, nil
}
