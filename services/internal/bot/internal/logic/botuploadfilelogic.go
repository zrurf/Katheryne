package logic

import (
	"context"
	"strconv"
	"time"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotUploadFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBotUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotUploadFileLogic {
	return &BotUploadFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BotUploadFileLogic) BotUploadFile(in *bot.BotUploadFileReq) (*bot.BotUploadFileResp, error) {
	fileID := strconv.FormatInt(time.Now().UnixNano(), 10)
	return &bot.BotUploadFileResp{
		FileId:   fileID,
		FileName: in.FileName,
		FileSize: in.FileSize,
		Url:      "",
		MimeType: in.ContentType,
	}, nil
}