package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mem/internal/dao"
	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConsolidateMemoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsolidateMemoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsolidateMemoriesLogic {
	return &ConsolidateMemoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ConsolidateMemories merges multiple older memories into summarized summaries.
// This is useful for periodic memory compression — turning many EVENT memories
// into a single SUMMARY memory.
func (l *ConsolidateMemoriesLogic) ConsolidateMemories(in *mem.ConsolidateMemoriesReq) (*mem.ConsolidateMemoriesResp, error) {
	targetType := int32(in.TargetType)
	if targetType == 0 {
		targetType = int32(mem.MemoryType_EVENT)
	}

	olderThan := in.OlderThan
	if olderThan <= 0 {
		olderThan = time.Now().Add(-24 * time.Hour).Unix()
	}

	maxItems := in.MaxItems
	if maxItems <= 0 {
		maxItems = 50
	}

	// Fetch memories to consolidate
	rows, _, err := l.svcCtx.Postgres.ListMemories(l.ctx, in.TenantId, in.TenantType, targetType, 1, maxItems)
	if err != nil {
		logx.Errorf("consolidate: list memories failed: %v", err)
		return nil, err
	}

	// Filter by older_than
	var toMerge []*dao.MemoryRow
	for _, row := range rows {
		if row.CreatedAt.Unix() < olderThan {
			toMerge = append(toMerge, row)
		}
	}

	if len(toMerge) == 0 {
		return &mem.ConsolidateMemoriesResp{}, nil
	}

	if in.DryRun {
		return &mem.ConsolidateMemoriesResp{ConsolidatedCount: int64(len(toMerge))}, nil
	}

	// Merge contents into a summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 记忆摘要 (合并 %d 条记录)\n\n", len(toMerge)))
	for _, row := range toMerge {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n",
			row.CreatedAt.Format("2006-01-02 15:04"), row.Content))
	}

	// Save as SUMMARY type memory
	summaryID := generateMemoryID()
	summaryRow := &dao.MemoryRow{
		MemoryID:   summaryID,
		TenantID:   in.TenantId,
		TenantType: in.TenantType,
		MemoryType: int32(mem.MemoryType_SUMMARY),
		Content:    sb.String(),
		Importance: 0.7,
		Entities:   []string{},
	}

	if err := l.svcCtx.Postgres.InsertMemory(l.ctx, summaryRow); err != nil {
		return nil, err
	}

	// Delete source memories
	var deleted int64
	for _, row := range toMerge {
		if err := l.svcCtx.Postgres.DeleteMemory(l.ctx, row.MemoryID); err == nil {
			deleted++
		}
		_ = l.svcCtx.Qdrant.DeletePoint(l.ctx, row.MemoryID)
	}

	logx.Infof("consolidate: merged %d memories into %s, deleted=%d",
		len(toMerge), summaryID, deleted)

	return &mem.ConsolidateMemoriesResp{
		ConsolidatedCount: int64(len(toMerge)),
		NewMemoryCount:    1,
		DeletedCount:      deleted,
	}, nil
}