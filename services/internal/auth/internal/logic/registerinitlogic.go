package logic

import (
	"context"

	"auth/auth"
	"auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type RegisterInitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterInitLogic {
	return &RegisterInitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterInitLogic) RegisterInit(in *auth.RegisterInitReq) (*auth.RegisterInitResp, error) {
	req, err := l.svcCtx.OpaqueSvc.GetServer().Deserialize.RegistrationRequest(in.RegistrationRequest)
	if err != nil {
		l.Logger.Errorf("Failed to deserialize registration request: %v", err)
		return &auth.RegisterInitResp{
			ServerPublicKey:      nil,
			RegistrationResponse: nil,
			CredentialIdentifier: nil,
		}, errors.New(104, "Request Error")
	}

	var credId = []byte(in.Phone)

	serverPubKey := l.svcCtx.OpaqueSvc.GetServerPublicKey()

	resp, err := l.svcCtx.OpaqueSvc.GetServer().RegistrationResponse(req, credId, nil)

	if err != nil {
		l.Logger.Errorf("Failed to generate registration response: %v", err)
		return nil, errors.New(-1, "Internal Server Error")
	}

	// 序列化并返回
	respBytes := resp.Serialize()

	return &auth.RegisterInitResp{
		ServerPublicKey:      serverPubKey,
		RegistrationResponse: respBytes,
		CredentialIdentifier: credId,
	}, nil
}
