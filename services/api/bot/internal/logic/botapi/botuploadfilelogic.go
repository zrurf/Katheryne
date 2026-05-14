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
	return &types.BotUploadFileResp{
		FileID:   "",
		FileName: "",
		FileSize: 0,
		URL:      "",
		MimeType: "",
	}, nil
}
