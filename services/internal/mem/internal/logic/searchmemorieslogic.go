package logic

import (
	"context"
	"sort"
	"time"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMemoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchMemoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMemoriesLogic {
	return &SearchMemoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchMemories performs multi-path recall combining:
//  1. Vector semantic search (Qdrant) - semantic similarity
//  2. Graph-based entity traversal (PostgreSQL recursive CTE)
//  3. Keyword full-text search (PostgreSQL FTS)
//
// Results are fused using weighted rank fusion and re-ranked.
func (l *SearchMemoriesLogic) SearchMemories(in *mem.SearchMemoriesReq) (*mem.SearchMemoriesResp, error) {
	start := time.Now()
	topK := in.TopK
	if topK <= 0 {
		topK = 10
	}

	// Default weights
	vw := in.VectorWeight
	if vw <= 0 {
		vw = 0.5
	}
	gw := in.GraphWeight
	if gw <= 0 {
		gw = 0.3
	}
	kw := in.KeywordWeight
	if kw <= 0 {
		kw = 0.2
	}

	// Build memory type filter for PG queries
	var typeFilter []int32
	for _, t := range in.TypeFilter {
		typeFilter = append(typeFilter, int32(t))
	}

	// Map to accumulate fusion scores: memoryID -> result
	resultMap := make(map[string]*mem.MemorySearchResult)
	var allResults []*mem.MemorySearchResult
	var meta mem.RecallMeta

	// ===== PATH 1: Vector Semantic Search (Qdrant) =====
	if vw > 0 {
		// In production, we'd use an embedding service to get the query vector.
		// For now, we use a zero vector placeholder and rely on keyword search.
		// Once an embedding service is integrated, this will return semantic matches.
		queryVector := make([]float32, l.svcCtx.Config.Qdrant.VectorDim)
		hits, err := l.svcCtx.Qdrant.SearchVector(l.ctx, queryVector, in.TenantId, uint64(topK*2))
		if err != nil {
			logx.Errorf("vector search failed: %v", err)
		} else {
			meta.VectorHits = int32(len(hits))
			for _, hit := range hits {
				if _, exists := resultMap[hit.PointID]; !exists {
					row, err := l.svcCtx.Postgres.GetMemory(l.ctx, hit.PointID)
					if err != nil || row == nil {
						continue
					}
					if in.MinImportance > 0 && row.Importance < in.MinImportance {
						continue
					}
					r := &mem.MemorySearchResult{
						Memory:      toProto(row),
						VectorScore: hit.Score,
					}
					resultMap[hit.PointID] = r
					allResults = append(allResults, r)
				} else {
					resultMap[hit.PointID].VectorScore = hit.Score
				}
			}
		}
	}

	// ===== PATH 2: Keyword Full-Text Search (PostgreSQL FTS) =====
	if kw > 0 {
		ftsRows, err := l.svcCtx.Postgres.FullTextSearch(l.ctx, in.TenantId, in.TenantType, in.Query, topK*3, typeFilter)
		if err != nil {
			logx.Errorf("FTS search failed: %v", err)
		} else {
			meta.KeywordHits = int32(len(ftsRows))
			for i, row := range ftsRows {
				if in.MinImportance > 0 && row.Importance < in.MinImportance {
					continue
				}
				if existing, exists := resultMap[row.MemoryID]; exists {
					// The FTS rank is proportional to result position
					existing.KeywordScore = float64(len(ftsRows)-i) / float64(len(ftsRows))
				} else {
					r := &mem.MemorySearchResult{
						Memory:       toProto(row),
						KeywordScore: float64(len(ftsRows)-i) / float64(len(ftsRows)),
					}
					resultMap[row.MemoryID] = r
					allResults = append(allResults, r)
				}
			}
		}
	}

	// ===== PATH 3: Graph Entity Traversal (PostgreSQL recursive CTE) =====
	if gw > 0 && len(in.FilterEntities) > 0 {
		nodes, _, err := l.svcCtx.Postgres.QueryGraphTraverse(l.ctx, in.TenantId, in.TenantType,
			in.FilterEntities, 2, int32(topK*2))
		if err != nil {
			logx.Errorf("graph traversal failed: %v", err)
		} else {
			meta.GraphHits = int32(len(nodes))
			for _, node := range nodes {
				// Search for memories mentioning this graph entity
				entityRows, _ := l.svcCtx.Postgres.FullTextSearch(l.ctx,
					in.TenantId, in.TenantType, node.Name, topK, typeFilter)
				for _, row := range entityRows {
					if in.MinImportance > 0 && row.Importance < in.MinImportance {
						continue
					}
					if existing, exists := resultMap[row.MemoryID]; exists {
						existing.GraphScore = maxDouble(existing.GraphScore, 0.5)
					} else {
						r := &mem.MemorySearchResult{
							Memory:     toProto(row),
							GraphScore: 0.5,
						}
						resultMap[row.MemoryID] = r
						allResults = append(allResults, r)
					}
				}
			}
		}
	}

	// ===== Fusion: Weighted Rank Fusion =====
	for _, r := range allResults {
		r.FusionScore = float64(vw)*r.VectorScore + float64(kw)*r.KeywordScore + float64(gw)*r.GraphScore
	}

	// Sort by fusion score descending
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].FusionScore > allResults[j].FusionScore
	})

	// Trim to top_k
	if int32(len(allResults)) > topK {
		allResults = allResults[:topK]
	}

	meta.FusedCount = int32(len(allResults))
	if len(allResults) > 0 {
		for _, r := range allResults {
			meta.AvgFusionScore += r.FusionScore
		}
		meta.AvgFusionScore /= float64(len(allResults))
	}

	latency := time.Since(start).Milliseconds()

	// Record access for all returned memories
	for _, r := range allResults {
		_ = l.svcCtx.Postgres.RecordAccess(l.ctx, r.Memory.MemoryId)
	}

	logx.Infof("search memories: query='%s' tenant=%s hits=%d latency=%dms",
		truncate(in.Query, 50), in.TenantId, len(allResults), latency)

	return &mem.SearchMemoriesResp{
		Results:    allResults,
		Total:      int32(len(allResults)),
		LatencyMs:  latency,
		RecallMeta: &meta,
	}, nil
}

func maxDouble(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
