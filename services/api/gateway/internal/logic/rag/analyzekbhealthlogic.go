package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnalyzeKBHealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnalyzeKBHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyzeKBHealthLogic {
	return &AnalyzeKBHealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnalyzeKBHealthLogic) AnalyzeKBHealth(req *types.AnalyzeKBHealthRequest) (resp *types.AnalyzeKBHealthResponse, err error) {
	result, err := l.svcCtx.RagRpc.AnalyzeKBHealth(l.ctx, &ragclient.AnalyzeKBHealthReq{
		KbId: req.KbID,
	})
	if err != nil {
		return nil, err
	}
	report := result.Report
	conflicts := make([]types.ConflictInfo, 0, len(report.Conflicts))
	for _, c := range report.Conflicts {
		conflicts = append(conflicts, types.ConflictInfo{
			Entity:            c.Entity,
			ConflictingChunks: c.ConflictingChunks,
			Description:       c.Description,
		})
	}
	gaps := make([]types.CoverageGapInfo, 0, len(report.Gaps))
	for _, g := range report.Gaps {
		gaps = append(gaps, types.CoverageGapInfo{
			Topic:      g.Topic,
			Suggestion: g.Suggestion,
		})
	}
	return &types.AnalyzeKBHealthResponse{Report: types.KBHealthInfo{
		KbID:           report.KbId,
		TotalDocs:      report.TotalDocs,
		TotalChunks:    report.TotalChunks,
		TotalEntities:  report.TotalEntities,
		TotalRelations: report.TotalRelations,
		AvgChunkSize:   float64(report.AvgChunkSize),
		CoverageRate:   float64(report.CoverageRate),
		ConflictRate:   float64(report.ConflictRate),
		HealthStatus:   report.HealthStatus,
		LastUpdatedAt:  report.LastUpdatedAt,
		Conflicts:      conflicts,
		Gaps:           gaps,
	}}, nil
}
