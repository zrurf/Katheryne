// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotUploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotUploadFileLogic {
	return &BotUploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotUploadFileLogic) BotUploadFile() (resp *types.BotUploadFileResp, err error) {
	// todo: add your logic here and delete this line

	return
}
