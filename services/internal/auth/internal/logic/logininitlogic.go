package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"auth/auth"
	"auth/internal/model"
	"auth/internal/svc"

	"github.com/bytedance/gopkg/util/xxhash3"
	"github.com/bytemare/opaque"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type LoginInitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginInitLogic {
	return &LoginInitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginInitLogic) LoginInit(in *auth.LoginInitReq) (*auth.LoginInitResp, error) {
	ke1, err := l.svcCtx.OpaqueSvc.GetServer().Deserialize.KE1(in.Ke1)
	if err != nil {
		l.Logger.Infof("Failed to deserialize KE1: %v", err)
		return nil, errors.New(104, "Request Error")
	}

	// 获取用户的 RegistrationRecord（opaque_record）
	_, recordBytes, err := l.svcCtx.UserDao.GetUserRecord(l.ctx, in.Phone)
	if err != nil {
		l.Logger.Infof("User not found", err)
		return nil, errors.New(21, "User not found")
	}

	// 反序列化 RegistrationRecord
	regRecord, err := l.svcCtx.OpaqueSvc.GetServer().Deserialize.RegistrationRecord(recordBytes)
	if err != nil {
		l.Logger.Infof("Failed to deserialize registration record: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}

	var credId = []byte(in.Phone)

	// 构造 ClientRecord
	clientRecord := &opaque.ClientRecord{
		CredentialIdentifier: credId,
		ClientIdentity:       credId,
		RegistrationRecord:   regRecord,
	}

	// 调用 LoginInit 得到 KE2
	ke2, output, err := l.svcCtx.OpaqueSvc.GetServer().GenerateKE2(ke1, clientRecord)
	if err != nil {
		l.Logger.Info("Server login init failed: %v", err)
		return nil, errors.New(11, "Login failed")
	}

	sessionId := strconv.FormatUint(xxhash3.HashString(fmt.Sprintf("%s_%x@%d", in.Phone, in.Ke1, time.Now().UnixMilli())), 16)

	l.svcCtx.SessionDao.SaveLoginSession(l.ctx, sessionId, &model.LoginSession{
		User: in.Phone,
		MAC:  output.ClientMAC,
	}, 60)

	return &auth.LoginInitResp{
		Ke2:       ke2.Serialize(),
		SessionId: sessionId,
	}, nil
}
