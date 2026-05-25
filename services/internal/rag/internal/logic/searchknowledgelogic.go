package logic

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"rag/internal/svc"
	"rag/rag/rag"

	pb "github.com/qdrant/go-client/qdrant"
	"github.com/zeromicro/go-zero/core/logx"
)

type SearchKnowledgeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchKnowledgeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchKnowledgeLogic {
	return &SearchKnowledgeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchKnowledge performs multi-path recall: vector + graph + keyword fusion.
func (l *SearchKnowledgeLogic) SearchKnowledge(in *rag.SearchKnowledgeReq) (*rag.SearchKnowledgeResp, error) {
	start := time.Now()

	topK := in.TopK
	if topK < 1 || topK > 100 {
		topK = 10
	}

	// Default weights
	vw := float64(in.VectorWeight)
	gw := float64(in.GraphWeight)
	kw := float64(in.KeywordWeight)
	if vw+gw+kw == 0 {
		vw, gw, kw = 0.5, 0.3, 0.2
	}
	totalW := vw + gw + kw
	vw /= totalW
	gw /= totalW
	kw /= totalW

	// Result collector: chunkID -> RecallItem
	resultMap := make(map[string]*rag.RecallItem)

	// Rerank strategy
	rerank := in.RerankStrategy
	if rerank == "" {
		rerank = "fusion"
	}

	// === Path 1: Vector search (Qdrant) ===
	if l.svcCtx.Qdrant != nil {
		queryVec := textToVector(in.Query, l.svcCtx.Config.Qdrant.VectorDim)

		scoredPoints, err := l.svcCtx.Qdrant.Search(l.ctx, in.KbId, queryVec, uint64(topK*3), nil)
		if err != nil {
			l.Errorf("qdrant search: %v", err)
		} else {
			for _, sp := range scoredPoints {
				chunkID := sp.Id.GetUuid()
				item := getOrCreate(resultMap, chunkID)
				item.VectorScore = float32(sp.Score)
				if sp.Payload != nil {
					if docID, ok := sp.Payload["doc_id"]; ok {
						item.DocId = docID.GetStringValue()
					}
					if docName, ok := sp.Payload["doc_name"]; ok {
						item.DocName = docName.GetStringValue()
					}
					if content, ok := sp.Payload["content"]; ok {
						item.Content = content.GetStringValue()
					}
					if chunkIdx, ok := sp.Payload["chunk_index"]; ok {
						item.ChunkIndex = int32(chunkIdx.GetIntegerValue())
					}
				}
			}
		}
	}

	// === Path 2: Graph traversal (HugeGraph) ===
	if l.svcCtx.HugeGraph != nil && gw > 0 {
		// Extract query entities
		queryEntities := extractKeywords(in.Query)
		graphScores := make(map[string]float64)

		for _, entity := range queryEntities {
			entityID := in.KbId + "_" + sanitizeEntityID(entity)
			verts, _, err := l.svcCtx.HugeGraph.GetSubGraph(l.ctx, entityID, 2, int(topK*2))
			if err != nil {
				continue
			}
			for _, v := range verts {
				// Extract chunk vertices
				if label, ok := v["label"].(string); ok && label == "chunk" {
					if props, ok := v["properties"].(map[string]interface{}); ok {
						if chunkID, ok := props["chunk_id"].(string); ok {
							graphScores[chunkID] += 0.5 // proximity score
						}
					}
				}
			}
		}

		for chunkID, score := range graphScores {
			item := getOrCreate(resultMap, chunkID)
			item.GraphScore = float32(math.Min(score, 1.0))
		}
	}

	// === Path 3: Keyword search (BM25-like) ===
	if kw > 0 {
		// Scroll through Qdrant points and do keyword matching
		if l.svcCtx.Qdrant != nil {
			queryTerms := strings.Fields(strings.ToLower(in.Query))
			var offset *pb.PointId
			for {
				resp, err := l.svcCtx.Qdrant.Scroll(l.ctx, in.KbId, offset, 500)
				if err != nil {
					break
				}
				for _, point := range resp.Result {
					if point.Payload == nil {
						continue
					}
					content := ""
					if c, ok := point.Payload["content"]; ok {
						content = strings.ToLower(c.GetStringValue())
					}
					if content == "" {
						continue
					}

					// Simple BM25-like scoring
					score := keywordScore(content, queryTerms)
					if score > 0 {
						chunkID := point.Id.GetUuid()
						item := getOrCreate(resultMap, chunkID)
						item.KeywordScore = float32(score)
					}
				}
				if resp.NextPageOffset == nil {
					break
				}
				offset = resp.NextPageOffset
			}
		}
	}

	// === Fusion & Ranking ===
	var items []*rag.RecallItem
	for _, item := range resultMap {
		item.FusionScore = float32(
			float64(item.VectorScore)*vw +
				float64(item.GraphScore)*gw +
				float64(item.KeywordScore)*kw,
		)
		items = append(items, item)
	}

	// Sort by fusion score
	sort.Slice(items, func(i, j int) bool {
		return items[i].FusionScore > items[j].FusionScore
	})

	// Apply rerank strategy
	switch rerank {
	case "mmr":
		items = mmrRerank(items, int(topK), 0.7)
	default: // fusion
		if len(items) > int(topK) {
			items = items[:topK]
		}
	}

	latency := time.Since(start).Milliseconds()

	// === Meta-cognition scoring ===
	metacog := assessMetaCognition(in.Query, items)

	return &rag.SearchKnowledgeResp{
		Items:     items,
		Total:     int64(len(items)),
		LatencyMs: latency,
		Metacog:   metacog,
	}, nil
}

func getOrCreate(m map[string]*rag.RecallItem, chunkID string) *rag.RecallItem {
	if item, ok := m[chunkID]; ok {
		return item
	}
	item := &rag.RecallItem{ChunkId: chunkID}
	m[chunkID] = item
	return item
}

// keywordScore computes a simple keyword match score.
func keywordScore(content string, terms []string) float64 {
	if len(terms) == 0 {
		return 0
	}
	hits := 0
	for _, term := range terms {
		if strings.Contains(content, term) {
			hits++
		}
	}
	return float64(hits) / float64(len(terms))
}

// mmrRerank applies Maximal Marginal Relevance for diversity.
func mmrRerank(items []*rag.RecallItem, k int, lambda float64) []*rag.RecallItem {
	if len(items) <= k {
		return items
	}

	selected := make([]*rag.RecallItem, 0, k)
	candidates := make([]*rag.RecallItem, len(items))
	copy(candidates, items)

	// Pick first (highest score)
	selected = append(selected, candidates[0])
	candidates = candidates[1:]

	for len(selected) < k && len(candidates) > 0 {
		bestIdx := 0
		bestScore := -1.0

		for i, c := range candidates {
			relevance := float64(c.FusionScore)
			maxSim := 0.0
			for _, s := range selected {
				sim := jaccardSim(c.Content, s.Content)
				if sim > maxSim {
					maxSim = sim
				}
			}
			mmr := lambda*relevance - (1-lambda)*maxSim
			if mmr > bestScore {
				bestScore = mmr
				bestIdx = i
			}
		}

		selected = append(selected, candidates[bestIdx])
		candidates = append(candidates[:bestIdx], candidates[bestIdx+1:]...)
	}

	return selected
}

// jaccardSim computes simple Jaccard similarity between two texts.
func jaccardSim(a, b string) float64 {
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	setA := make(map[string]bool)
	for _, w := range wordsA {
		setA[w] = true
	}
	setB := make(map[string]bool)
	for _, w := range wordsB {
		setB[w] = true
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// assessMetaCognition generates meta-cognition scores for search results.
func assessMetaCognition(query string, items []*rag.RecallItem) *rag.MetaCognitionScore {
	if len(items) == 0 {
		return &rag.MetaCognitionScore{
			OverallConfidence: 0,
			Warnings:          []string{"no relevant results found"},
		}
	}

	// Overall confidence based on top score
	topScore := float64(items[0].FusionScore)
	confidence := math.Min(topScore*2, 1.0) // Scale up

	// Coverage: how many unique docs were retrieved
	docSet := make(map[string]bool)
	for _, item := range items {
		if item.DocId != "" {
			docSet[item.DocId] = true
		}
	}
	coverage := math.Min(float64(len(docSet))/3.0, 1.0)

	// Relevance: average score of top results
	var sumScore float64
	for _, item := range items {
		sumScore += float64(item.FusionScore)
	}
	relevance := sumScore / float64(len(items))

	// Consistency: check for conflicting scores (high variance = potential conflict)
	var variance float64
	mean := relevance
	for _, item := range items {
		diff := float64(item.FusionScore) - mean
		variance += diff * diff
	}
	variance /= float64(len(items))
	consistency := 1.0 - math.Min(math.Sqrt(variance)*2, 1.0)

	// Freshness: placeholder (in production, check document timestamps)
	freshness := 0.8

	// Warnings
	var warnings []string
	if confidence < 0.3 {
		warnings = append(warnings, "low confidence in search results")
	}
	if consistency < 0.5 {
		warnings = append(warnings, "potential knowledge conflicts detected")
	}
	if coverage < 0.3 {
		warnings = append(warnings, "limited knowledge coverage")
	}

	// Citations
	var citations []*rag.Citation
	for _, item := range items {
		if item.DocName != "" {
			citations = append(citations, &rag.Citation{
				DocName:      item.DocName,
				ChunkId:      item.ChunkId,
				ChunkIndex:   item.ChunkIndex,
				Contribution: item.FusionScore,
				Excerpt:      truncateText(item.Content, 200),
			})
		}
	}

	return &rag.MetaCognitionScore{
		OverallConfidence: float32(confidence),
		CoverageScore:     float32(coverage),
		RelevanceScore:    float32(relevance),
		ConsistencyScore:  float32(consistency),
		FreshnessScore:    float32(freshness),
		Warnings:          warnings,
		Citations:         citations,
	}
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}
