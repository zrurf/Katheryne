package logic

import (
	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
)

type RevokeBotKBAccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevokeBotKBAccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeBotKBAccessLogic {
	return &RevokeBotKBAccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RevokeBotKBAccessLogic) RevokeBotKBAccess(in *rag.RevokeBotKBAccessReq) (*rag.RevokeBotKBAccessResp, error) {
	if err := l.svcCtx.Storage.RevokeBotKBAccess(l.ctx, in.Uid, in.BotId, in.ConvId); err != nil {
		return nil, err
	}
	return &rag.RevokeBotKBAccessResp{}, nil
}