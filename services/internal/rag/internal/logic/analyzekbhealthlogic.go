package logic

import (
	"context"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnalyzeKBHealthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAnalyzeKBHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyzeKBHealthLogic {
	return &AnalyzeKBHealthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AnalyzeKBHealthLogic) AnalyzeKBHealth(in *rag.AnalyzeKBHealthReq) (*rag.AnalyzeKBHealthResp, error) {
	kb, err := l.svcCtx.Storage.GetKB(l.ctx, in.KbId)
	if err != nil {
		return nil, errKBNotFound
	}

	report := &rag.KBHealthReport{
		KbId:       kb.KbID,
		TotalDocs:  kb.DocCount,
		TotalChunks: kb.ChunkCount,
	}

	// Get graph statistics
	if l.svcCtx.HugeGraph != nil {
		entityCount, err := l.svcCtx.HugeGraph.GetKBEntityCount(l.ctx, in.KbId)
		if err == nil {
			report.TotalEntities = entityCount
		}
		relationCount, err := l.svcCtx.HugeGraph.GetKBRelationCount(l.ctx, in.KbId)
		if err == nil {
			report.TotalRelations = relationCount
		}
	}

	// Average chunk size
	if kb.ChunkCount > 0 {
		report.AvgChunkSize = float32(kb.TotalSize) / float32(kb.ChunkCount)
	}

	// Entity coverage rate: entities per chunk ratio
	if kb.ChunkCount > 0 && report.TotalEntities > 0 {
		report.CoverageRate = float32(report.TotalEntities) / float32(kb.ChunkCount)
	}

	// Conflict detection: placeholder (in production, use NLP to compare facts)
	report.ConflictRate = 0.05 // placeholder
	report.Conflicts = []*rag.ConflictItem{}

	// Coverage gaps: check if entity density is low
	if report.CoverageRate < 1.0 {
		report.Gaps = append(report.Gaps, &rag.CoverageGap{
			Topic:      "general",
			Suggestion: "Consider adding more diverse documents to improve entity coverage",
		})
	}

	// Determine health status
	report.LastUpdatedAt = kb.UpdatedAt.UnixMilli()
	switch {
	case kb.DocCount == 0:
		report.HealthStatus = "STALE"
	case report.CoverageRate < 0.5:
		report.HealthStatus = "NEEDS_REVIEW"
	default:
		report.HealthStatus = "HEALTHY"
	}

	return &rag.AnalyzeKBHealthResp{Report: report}, nil
}