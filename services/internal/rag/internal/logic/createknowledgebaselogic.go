package logic

import (
	"context"
	"crypto/rand"
	"fmt"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateKnowledgeBaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateKnowledgeBaseLogic {
	return &CreateKnowledgeBaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateKnowledgeBaseLogic) CreateKnowledgeBase(in *rag.CreateKBReq) (*rag.CreateKBResp, error) {
	kbID := generateID("kb")
	sourceType := in.SourceType
	if sourceType == "" {
		sourceType = "PLATFORM"
	}
	sourceConfig := in.SourceConfig
	if sourceConfig == "" {
		sourceConfig = "{}"
	}
	if err := l.svcCtx.Storage.CreateKB(l.ctx, kbID, in.Name, in.Description, sourceType, sourceConfig, in.OwnerUid); err != nil {
		return nil, fmt.Errorf("create kb: %w", err)
	}

	// Create Qdrant collection for this knowledge base
	if l.svcCtx.Qdrant != nil {
		if err := l.svcCtx.Qdrant.EnsureCollection(l.ctx, kbID); err != nil {
			l.Errorf("qdrant ensure collection %s: %v", kbID, err)
			// Non-fatal: can create later on first indexing
		}
	}

	// If it's an external KB, create an initial sync record
	if sourceType != "PLATFORM" {
		if err := l.svcCtx.Storage.CreateExternalSync(l.ctx, kbID, sourceType, sourceConfig); err != nil {
			l.Errorf("create external sync record for %s: %v", kbID, err)
		}
	}

	return &rag.CreateKBResp{KbId: kbID}, nil
}

func generateID(prefix string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%x", prefix, b)
}
