// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegenerateCredentialLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegenerateCredentialLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegenerateCredentialLogic {
	return &RegenerateCredentialLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegenerateCredentialLogic) RegenerateCredential(req *types.RegenerateCredentialReq) (resp *types.RegenerateCredentialResp, err error) {
	// todo: add your logic here and delete this line

	return
}
