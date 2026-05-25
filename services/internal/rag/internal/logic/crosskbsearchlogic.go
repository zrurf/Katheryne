package logic

import (
	"context"
	"time"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type CrossKBSearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCrossKBSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CrossKBSearchLogic {
	return &CrossKBSearchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CrossKBSearchLogic) CrossKBSearch(in *rag.CrossKBSearchReq) (*rag.CrossKBSearchResp, error) {
	start := time.Now()
	topK := in.TopK
	if topK < 1 || topK > 100 {
		topK = 10
	}

	var allItems []*rag.RecallItem
	kbCounts := make(map[string]int64)
	var allMetacog []*rag.MetaCognitionScore

	for _, kbID := range in.KbIds {
		// Delegate to single-KB search
		searchLogic := NewSearchKnowledgeLogic(l.ctx, l.svcCtx)
		resp, err := searchLogic.SearchKnowledge(&rag.SearchKnowledgeReq{
			KbId:          kbID,
			Query:         in.Query,
			TopK:          topK,
			VectorWeight:  in.VectorWeight,
			GraphWeight:   in.GraphWeight,
			KeywordWeight: in.KeywordWeight,
		})
		if err != nil {
			l.Errorf("search kb %s: %v", kbID, err)
			continue
		}

		allItems = append(allItems, resp.Items...)
		kbCounts[kbID] = resp.Total
		if resp.Metacog != nil {
			allMetacog = append(allMetacog, resp.Metacog)
		}
	}

	// Re-rank across all KBs
	// Sort by fusion score (already sorted per KB, but merge sort)
	sortItemsByFusion(allItems)
	if len(allItems) > int(topK) {
		allItems = allItems[:topK]
	}

	// Aggregate meta-cognition
	metacog := aggregateMetaCognition(allMetacog, allItems)

	return &rag.CrossKBSearchResp{
		Items:     allItems,
		KbCounts:  kbCounts,
		LatencyMs: time.Since(start).Milliseconds(),
		Metacog:   metacog,
	}, nil
}

func sortItemsByFusion(items []*rag.RecallItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].FusionScore > items[i].FusionScore {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func aggregateMetaCognition(scores []*rag.MetaCognitionScore, items []*rag.RecallItem) *rag.MetaCognitionScore {
	if len(scores) == 0 {
		return assessMetaCognition("", items)
	}

	var totalConfidence, totalCoverage, totalRelevance, totalConsistency, totalFreshness float64
	for _, s := range scores {
		totalConfidence += float64(s.OverallConfidence)
		totalCoverage += float64(s.CoverageScore)
		totalRelevance += float64(s.RelevanceScore)
		totalConsistency += float64(s.ConsistencyScore)
		totalFreshness += float64(s.FreshnessScore)
	}
	n := float64(len(scores))

	var allWarnings []string
	for _, s := range scores {
		allWarnings = append(allWarnings, s.Warnings...)
	}
	var allCitations []*rag.Citation
	for _, s := range scores {
		allCitations = append(allCitations, s.Citations...)
	}

	return &rag.MetaCognitionScore{
		OverallConfidence: float32(totalConfidence / n),
		CoverageScore:     float32(totalCoverage / n),
		RelevanceScore:    float32(totalRelevance / n),
		ConsistencyScore:  float32(totalConsistency / n),
		FreshnessScore:    float32(totalFreshness / n),
		Warnings:          allWarnings,
		Citations:         allCitations,
	}
}