package logic

import (
	"context"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExtractMemoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExtractMemoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExtractMemoriesLogic {
	return &ExtractMemoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ExtractMemories analyzes a text block and returns extracted memory items.
// In production, this would call an LLM to extract entities, facts, preferences, etc.
// Currently, it returns the text as a simple fact memory.
func (l *ExtractMemoriesLogic) ExtractMemories(in *mem.ExtractMemoriesReq) (*mem.ExtractMemoriesResp, error) {
	if in.Text == "" {
		return &mem.ExtractMemoriesResp{}, nil
	}

	// Simple extraction: treat the text as a FACT memory
	// In production, call an LLM to extract structured memories
	items := []*mem.ExtractedMemory{
		{
			Type:       mem.MemoryType_FACT,
			Content:    in.Text,
			Importance: 0.3,
			Entities:   []string{},
		},
	}

	logx.Infof("extract: extracted %d memories for tenant=%s", len(items), in.TenantId)
	return &mem.ExtractMemoriesResp{Items: items}, nil
}