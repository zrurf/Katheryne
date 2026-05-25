package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssessRetrievalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssessRetrievalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssessRetrievalLogic {
	return &AssessRetrievalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssessRetrievalLogic) AssessRetrieval(req *types.AssessRetrievalRequest) (resp *types.AssessRetrievalResponse, err error) {
	// Convert gateway types to RPC types
	results := make([]*ragclient.RecallItem, 0, len(req.Results))
	for _, r := range req.Results {
		results = append(results, &ragclient.RecallItem{
			ChunkId:      r.ChunkID,
			DocId:        r.DocID,
			DocName:      r.DocName,
			Content:      r.Content,
			VectorScore:  float32(r.VectorScore),
			GraphScore:   float32(r.GraphScore),
			KeywordScore: float32(r.KeywordScore),
			FusionScore:  float32(r.FusionScore),
			Entities:     r.Entities,
			Relations:    r.Relations,
			ChunkIndex:   r.ChunkIndex,
		})
	}
	result, err := l.svcCtx.RagRpc.AssessRetrieval(l.ctx, &ragclient.AssessRetrievalReq{
		Query:   req.Query,
		Results: results,
	})
	if err != nil {
		return nil, err
	}
	score := result.Score
	return &types.AssessRetrievalResponse{Score: types.MetaCogInfo{
		OverallConfidence: float64(score.OverallConfidence),
		CoverageScore:     float64(score.CoverageScore),
		RelevanceScore:    float64(score.RelevanceScore),
		ConsistencyScore:  float64(score.ConsistencyScore),
		FreshnessScore:    float64(score.FreshnessScore),
		Warnings:          score.Warnings,
	}}, nil
}
