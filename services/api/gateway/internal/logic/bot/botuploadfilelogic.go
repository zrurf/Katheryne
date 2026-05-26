package bot

import (
	"context"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

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

func (l *BotUploadFileLogic) BotUploadFile(req *types.BotUploadFileReq) (resp *types.BotUploadFileResp, err error) {
	result, err := l.svcCtx.BotRpc.BotUploadFile(l.ctx, &botclient.BotUploadFileReq{})
	if err != nil {
		return nil, err
	}
	return &types.BotUploadFileResp{
		FileId:   result.FileId,
		FileName: result.FileName,
		FileSize: result.FileSize,
		Url:      result.Url,
		MimeType: result.MimeType,
	}, nil
}
