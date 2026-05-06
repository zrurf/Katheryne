package logic

import (
	"context"

	"auth/auth"
	"auth/internal/dao"
	"auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type RegisterFinalizeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext

	logx.Logger
}

func NewRegisterFinalizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterFinalizeLogic {
	return &RegisterFinalizeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterFinalizeLogic) RegisterFinalize(in *auth.RegisterFinalizeReq) (*auth.RegisterFinalizeResp, error) {
	// 反序列化验证格式，提前发现客户端错误
	record, err := l.svcCtx.OpaqueSvc.GetServer().Deserialize.RegistrationRecord(in.RegistrationRecord)
	if err != nil {
		l.Logger.Infof("Failed to deserialize registration record: %v", err)
		return &auth.RegisterFinalizeResp{
			Result: false,
			Reason: "invalid registration record",
		}, errors.New(104, "Request Error")
	}
	_ = record // 回收

	PhoneExists, err := l.svcCtx.UserDao.PhoneExists(l.ctx, in.Phone)

	if PhoneExists || err != nil {
		l.Logger.Infof("Phone %s already exists", in.Phone)
		return &auth.RegisterFinalizeResp{
			Result: false,
			Reason: "phone already exists",
		}, errors.New(23, "Phone already exists")
	}

	var uid int64
	var exists bool

	for retry := 0; retry < l.svcCtx.Config.MaxUidRetries; retry++ {
		id, err := l.svcCtx.SonyFlake.NextID()
		if err != nil {
			l.Logger.Errorf("Failed to generate unique UID: %v", err)
			return &auth.RegisterFinalizeResp{
				Result: false,
				Reason: "Internal Server Error",
			}, errors.New(-1, "Internal Server Error")
		}
		uid = int64(id)

		exists, err = l.svcCtx.UserDao.UidExists(l.ctx, uid)
		if err != nil {
			l.Logger.Errorf("Failed to check if UID exists: %v", err)
			return &auth.RegisterFinalizeResp{
				Result: false,
				Reason: "Internal Server Error",
			}, errors.New(-1, "Internal Server Error")
		}
		if !exists {
			break
		}
	}

	if exists {
		l.Logger.Errorf("Failed to generate unique UID after %d retries", l.svcCtx.Config.MaxUidRetries)
		return &auth.RegisterFinalizeResp{
			Result: false,
			Reason: "Internal Server Error",
		}, errors.New(-1, "Internal Server Error")
	}

	// 将注册记录存到数据库 opaque_record 字段
	if err := l.svcCtx.UserDao.SaveUserRecord(l.ctx, int64(uid), in.Phone, in.Name, in.RegistrationRecord); err != nil {
		l.Logger.Errorf("Failed to save user opaque record: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}

	var language string = "zh-CN"
	var bTrue = true
	var bFalse = false

	l.svcCtx.UserDao.UpsertUserConfig(l.ctx, &dao.UpsertUserConfig{
		UID:             uid,
		Language:        &language,
		MsgNotification: &bTrue,
		SoundEnabled:    &bTrue,
		AutoTranslate:   &bFalse,
		TranslateTarget: &language,
		ContentFilter:   &bFalse,
		Enable2FA:       &bFalse,
		TOTPSecret:      nil,
		TOTPBackupCodes: nil,
	})

	return &auth.RegisterFinalizeResp{
		Result: true,
		Reason: "",
	}, nil
}
