package logic

import (
	"context"

	"user/internal/dao"
	"user/internal/svc"
	"user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSettingsLogic {
	return &GetUserSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserSettingsLogic) GetUserSettings(in *user.GetUserSettingsReq) (*user.GetUserSettingsResp, error) {
	if in.Uid <= 0 {
		return &user.GetUserSettingsResp{Settings: defaultSettings()}, nil
	}

	// 先查缓存
	cfg, err := l.svcCtx.RedisDao.GetUserSettingsCache(l.ctx, in.Uid)
	if err == nil && cfg != nil {
		return &user.GetUserSettingsResp{Settings: daoConfigToProto(cfg)}, nil
	}

	// 再查数据库
	cfg, err = l.svcCtx.UserDao.GetUserConfig(l.ctx, in.Uid)
	if err != nil {
		l.Errorf("GetUserConfig error: %v", err)
		return nil, err
	}
	if cfg == nil {
		return &user.GetUserSettingsResp{Settings: defaultSettings()}, nil
	}

	// 写入缓存
	go func() {
		_ = l.svcCtx.RedisDao.SetUserSettingsCache(context.Background(), in.Uid, cfg)
	}()

	return &user.GetUserSettingsResp{Settings: daoConfigToProto(cfg)}, nil
}

func defaultSettings() *user.UserSettings {
	return &user.UserSettings{
		Language:        "zh-CN",
		MsgNotification: true,
		SoundEnabled:    true,
		AutoTranslate:   false,
		TranslateTarget: "",
		ContentFilter:   true,
	}
}

func daoConfigToProto(cfg *dao.UserConfig) *user.UserSettings {
	if cfg == nil {
		return defaultSettings()
	}
	var target string
	if cfg.TranslateTarget != nil {
		target = *cfg.TranslateTarget
	}
	return &user.UserSettings{
		Language:        cfg.Language,
		MsgNotification: cfg.MsgNotification,
		SoundEnabled:    cfg.SoundEnabled,
		AutoTranslate:   cfg.AutoTranslate,
		TranslateTarget: target,
		ContentFilter:   cfg.ContentFilter,
	}
}
