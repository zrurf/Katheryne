package logic

import (
	"context"

	"conversation/conversation"
	"conversation/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrCreateSingleConvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrCreateSingleConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateSingleConvLogic {
	return &GetOrCreateSingleConvLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrCreateSingleConvLogic) GetOrCreateSingleConv(in *conversation.GetOrCreateSingleConvReq) (*conversation.GetOrCreateSingleConvResp, error) {
	// todo: add your logic here and delete this line

	return &conversation.GetOrCreateSingleConvResp{}, nil
}
