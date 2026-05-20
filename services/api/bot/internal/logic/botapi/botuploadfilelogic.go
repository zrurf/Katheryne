package botapi

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"bot/internal/middleware"
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
	auth, err := middleware.GetBotAuth(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	_ = auth

	fileID := strconv.FormatInt(time.Now().UnixNano(), 10)
	return &types.BotUploadFileResp{
		FileID:   fileID,
		FileName: "uploaded",
		FileSize: 0,
		URL:      "",
		MimeType: "application/octet-stream",
	}, nil
}