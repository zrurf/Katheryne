package logic

import (
	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
)

type GrantBotKBAccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGrantBotKBAccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantBotKBAccessLogic {
	return &GrantBotKBAccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GrantBotKBAccessLogic) GrantBotKBAccess(in *rag.GrantBotKBAccessReq) (*rag.GrantBotKBAccessResp, error) {
	if err := l.svcCtx.Storage.GrantBotKBAccess(l.ctx, in.Uid, in.BotId, in.ConvId); err != nil {
		return nil, err
	}
	return &rag.GrantBotKBAccessResp{}, nil
}