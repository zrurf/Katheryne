package logic

import (
	"context"
	"fmt"
	"os"

	"rag/internal/dao"
	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerExternalSyncLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTriggerExternalSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerExternalSyncLogic {
	return &TriggerExternalSyncLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TriggerExternalSyncLogic) TriggerExternalSync(in *rag.TriggerExternalSyncReq) (*rag.TriggerExternalSyncResp, error) {
	// Verify the KB exists
	kb, err := l.svcCtx.Storage.GetKB(l.ctx, in.KbId)
	if err != nil {
		return nil, fmt.Errorf("kb not found: %w", err)
	}

	if kb.SourceType == "PLATFORM" {
		return nil, fmt.Errorf("kb %s is a platform KB, not an external source", in.KbId)
	}

	// Create or update sync record
	if err := l.svcCtx.Storage.CreateExternalSync(l.ctx, in.KbId, kb.SourceType, kb.SourceConfig); err != nil {
		return nil, fmt.Errorf("create sync record: %w", err)
	}

	syncID := "sync_" + in.KbId

	// Kick off the actual sync asynchronously
	go l.runSync(syncID, kb)

	return &rag.TriggerExternalSyncResp{
		SyncId: syncID,
		Status: "TRIGGERED",
	}, nil
}

func (l *TriggerExternalSyncLogic) runSync(syncID string, kb *dao.KnowledgeBaseRow) {
	ctx := context.Background()

	if err := l.svcCtx.Storage.UpdateExternalSyncStatus(ctx, syncID, "SYNCING", ""); err != nil {
		l.Errorf("update sync status to SYNCING: %v", err)
	}

	var syncErr error
	switch kb.SourceType {
	case "FEISHU":
		syncErr = l.syncFeishu(ctx, kb)
	case "NOTION":
		syncErr = l.syncNotion(ctx, kb)
	default:
		syncErr = fmt.Errorf("unsupported source type: %s", kb.SourceType)
	}

	if syncErr != nil {
		l.Errorf("sync %s failed: %v", syncID, syncErr)
		l.svcCtx.Storage.UpdateExternalSyncStatus(ctx, syncID, "FAILED", syncErr.Error())
	} else {
		l.svcCtx.Storage.UpdateExternalSyncStatus(ctx, syncID, "SYNCED", "")
	}
}

func (l *TriggerExternalSyncLogic) syncFeishu(ctx context.Context, kb *dao.KnowledgeBaseRow) error {
	l.Infof("Starting Feishu sync for kb %s", kb.KbID)

	return SyncFeishuDocuments(ctx, kb, func(docName, contentType string, data []byte) (string, error) {
		docID := generateID("doc_fs")
		if err := l.svcCtx.Storage.CreateDocument(ctx, docID, kb.KbID,
			docName, contentType, int64(len(data)), ""); err != nil {
			return "", fmt.Errorf("create document: %w", err)
		}

		// Write data to temp file for streaming processing
		tmpFile, err := os.CreateTemp("", "rag-fs-*")
		if err != nil {
			return "", fmt.Errorf("create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		if _, err := tmpFile.Write(data); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return "", fmt.Errorf("write temp file: %w", err)
		}
		tmpFile.Close()

		go ProcessAndIndexDocument(context.Background(), l.svcCtx, docID, kb.KbID, docName, contentType, tmpPath)
		return docID, nil
	})
}

func (l *TriggerExternalSyncLogic) syncNotion(ctx context.Context, kb *dao.KnowledgeBaseRow) error {
	// TODO: Implement Notion document sync
	// - Parse source_config JSON for Notion API token and database/page IDs
	// - Fetch pages/blocks via Notion API
	// - Convert to plain text and chunk
	// - Index into Qdrant and HugeGraph
	l.Infof("Notion sync for kb %s is not yet implemented", kb.KbID)
	return nil
}
