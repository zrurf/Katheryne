package logic

import (
	"context"

	"mem/internal/dao"
	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchAddMemoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchAddMemoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchAddMemoryLogic {
	return &BatchAddMemoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchAddMemoryLogic) BatchAddMemory(in *mem.BatchAddMemoryReq) (*mem.BatchAddMemoryResp, error) {
	var ids []string

	for _, item := range in.Items {
		memoryID := generateMemoryID()

		importance := item.Importance
		if importance <= 0 {
			importance = 0.5
		}

		metadata := item.Metadata
		if metadata == nil {
			metadata = make(map[string]string)
		}

		memoryType := int32(item.Type)
		if memoryType == 0 {
			memoryType = int32(mem.MemoryType_FACT)
		}

		row := &dao.MemoryRow{
			MemoryID:   memoryID,
			TenantID:   item.TenantId,
			TenantType: item.TenantType,
			MemoryType: memoryType,
			Content:    item.Content,
			Importance: importance,
			Entities:   []string{},
			Metadata:   metadata,
			ExpiresAt:  computeExpiry(item.TtlSeconds),
		}

		if err := l.svcCtx.Postgres.InsertMemory(l.ctx, row); err != nil {
			logx.Errorf("batch insert memory failed: id=%s err=%v", memoryID, err)
			continue
		}
		_ = l.svcCtx.Qdrant.UpsertPoint(l.ctx, memoryID,
			make([]float32, l.svcCtx.Config.Qdrant.VectorDim),
			item.TenantId, item.TenantType)
		ids = append(ids, memoryID)
	}

	logx.Infof("batch added %d memories", len(ids))
	return &mem.BatchAddMemoryResp{MemoryIds: ids}, nil
}