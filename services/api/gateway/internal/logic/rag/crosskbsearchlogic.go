package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CrossKBSearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCrossKBSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CrossKBSearchLogic {
	return &CrossKBSearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CrossKBSearchLogic) CrossKBSearch(req *types.CrossKBSearchRequest) (resp *types.CrossKBSearchResponse, err error) {
	result, err := l.svcCtx.RagRpc.CrossKBSearch(l.ctx, &ragclient.CrossKBSearchReq{
		KbIds:         req.KbIDs,
		Query:         req.Query,
		TopK:          int32(req.TopK),
		VectorWeight:  float32(req.VectorWeight),
		GraphWeight:   float32(req.GraphWeight),
		KeywordWeight: float32(req.KeywordWeight),
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.RecallItemInfo, 0, len(result.Items))
	for _, r := range result.Items {
		items = append(items, types.RecallItemInfo{
			ChunkID:      r.ChunkId,
			DocID:        r.DocId,
			DocName:      r.DocName,
			Content:      r.Content,
			VectorScore:  float64(r.VectorScore),
			GraphScore:   float64(r.GraphScore),
			KeywordScore: float64(r.KeywordScore),
			FusionScore:  float64(r.FusionScore),
			Entities:     r.Entities,
			Relations:    r.Relations,
			ChunkIndex:   r.ChunkIndex,
		})
	}
	resp = &types.CrossKBSearchResponse{
		Items:     items,
		KbCounts:  result.KbCounts,
		LatencyMs: result.LatencyMs,
	}
	if result.Metacog != nil {
		resp.Metacog = &types.MetaCogInfo{
			OverallConfidence: float64(result.Metacog.OverallConfidence),
			CoverageScore:     float64(result.Metacog.CoverageScore),
			RelevanceScore:    float64(result.Metacog.RelevanceScore),
			ConsistencyScore:  float64(result.Metacog.ConsistencyScore),
			FreshnessScore:    float64(result.Metacog.FreshnessScore),
			Warnings:          result.Metacog.Warnings,
		}
	}
	return resp, nil
}
