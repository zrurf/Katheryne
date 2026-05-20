package developer

import (
	"context"
	"fmt"

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
	uid := l.ctx.Value("uid").(int64)

	if err := l.svcCtx.BotDao.CheckBotOwnership(l.ctx, req.BotID, uid); err != nil {
		return nil, fmt.Errorf("bot not found or not authorized")
	}

	clientID, clientSecret, err := l.svcCtx.BotDao.RegenerateCredential(l.ctx, req.BotID)
	if err != nil {
		return nil, err
	}

	return &types.RegenerateCredentialResp{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}