package logic

import (
	"context"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssessRetrievalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssessRetrievalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssessRetrievalLogic {
	return &AssessRetrievalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssessRetrievalLogic) AssessRetrieval(in *rag.AssessRetrievalReq) (*rag.AssessRetrievalResp, error) {
	score := assessMetaCognition(in.Query, in.Results)
	return &rag.AssessRetrievalResp{Score: score}, nil
}