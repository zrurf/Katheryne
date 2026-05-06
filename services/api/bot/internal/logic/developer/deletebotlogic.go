// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBotLogic {
	return &DeleteBotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBotLogic) DeleteBot(req *types.DeleteBotReq) (resp *types.DeleteBotResp, err error) {
	// todo: add your logic here and delete this line

	return
}
