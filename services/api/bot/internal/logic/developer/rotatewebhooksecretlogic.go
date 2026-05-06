// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RotateWebhookSecretLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRotateWebhookSecretLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RotateWebhookSecretLogic {
	return &RotateWebhookSecretLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RotateWebhookSecretLogic) RotateWebhookSecret(req *types.RotateWebhookSecretReq) (resp *types.RotateWebhookSecretResp, err error) {
	// todo: add your logic here and delete this line

	return
}
