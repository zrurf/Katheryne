// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBotLogic {
	return &GetBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBotLogic) GetBot(req *types.GetBotReq) (resp *types.GetBotResp, err error) {
	// todo: add your logic here and delete this line

	return
}
