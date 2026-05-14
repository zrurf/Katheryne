package developer

import (
	"context"
	"encoding/json"
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
	data, err := l.svcCtx.Redis.HGet(l.ctx, "bots", fmt.Sprintf("%d", req.BotID)).Result()
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	clientID := "bot_" + randomHex(16)
	clientSecret := randomHex(32)

	var bot map[string]interface{}
	json.Unmarshal([]byte(data), &bot)
	bot["client_id"] = clientID
	bot["client_secret"] = clientSecret

	data2, _ := json.Marshal(bot)
	l.svcCtx.Redis.HSet(l.ctx, "bots", fmt.Sprintf("%d", req.BotID), data2)

	return &types.RegenerateCredentialResp{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}
