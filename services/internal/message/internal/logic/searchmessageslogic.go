package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMessagesLogic {
	return &SearchMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchMessagesLogic) SearchMessages(in *message.SearchMessagesReq) (*message.SearchMessagesResp, error) {
	// todo: add your logic here and delete this line

	return &message.SearchMessagesResp{}, nil
}
