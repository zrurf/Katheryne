package logic

import (
	"context"

	"mem/internal/dao"
	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddMemoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddMemoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddMemoryLogic {
	return &AddMemoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddMemoryLogic) AddMemory(in *mem.AddMemoryReq) (*mem.AddMemoryResp, error) {
	memoryID := generateMemoryID()

	importance := in.Importance
	if importance <= 0 {
		importance = 0.5
	}

	metadata := in.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	memoryType := int32(in.Type)
	if memoryType == 0 {
		memoryType = int32(mem.MemoryType_FACT)
	}

	row := &dao.MemoryRow{
		MemoryID:   memoryID,
		TenantID:   in.TenantId,
		TenantType: in.TenantType,
		MemoryType: memoryType,
		Content:    in.Content,
		Importance: importance,
		Entities:   []string{},
		Metadata:   metadata,
		ExpiresAt:  computeExpiry(in.TtlSeconds),
	}

	if err := l.svcCtx.Postgres.InsertMemory(l.ctx, row); err != nil {
		logx.Errorf("insert memory failed: %v", err)
		return nil, err
	}

	// For vector search, we store the content as a vector point
	// In production, you'd call an embedding service here.
	// For now, use a placeholder zero vector (will be replaced when embedding service is integrated).
	_ = l.svcCtx.Qdrant.UpsertPoint(l.ctx, memoryID, make([]float32, l.svcCtx.Config.Qdrant.VectorDim),
		in.TenantId, in.TenantType)

	logx.Infof("memory added: id=%s tenant=%s type=%d", memoryID, in.TenantId, memoryType)
	return &mem.AddMemoryResp{MemoryId: memoryID}, nil
}