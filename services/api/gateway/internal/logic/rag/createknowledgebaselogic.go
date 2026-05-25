package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateKnowledgeBaseLogic {
	return &CreateKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateKnowledgeBaseLogic) CreateKnowledgeBase(req *types.CreateKBRequest) (resp *types.CreateKBResponse, err error) {
	uid := l.ctx.Value("uid").(int64)
	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = "PLATFORM"
	}
	sourceConfig := req.SourceConfig
	if sourceConfig == "" {
		sourceConfig = "{}"
	}
	result, err := l.svcCtx.RagRpc.CreateKnowledgeBase(l.ctx, &ragclient.CreateKBReq{
		OwnerUid:     uid,
		Name:         req.Name,
		Description:  req.Description,
		SourceType:   sourceType,
		SourceConfig: sourceConfig,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateKBResponse{KbID: result.KbId}, nil
}