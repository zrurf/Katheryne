package logic

import (
	"context"
	"fmt"
	"io"
	"os"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// maxFileSize is the maximum upload size (50MB).
	maxFileSize = 50 * 1024 * 1024
	// tempDirPattern for staging uploaded files before processing.
	tempDirPattern = "rag-upload-*"
)

type UploadDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadDocumentLogic {
	return &UploadDocumentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UploadDocument handles streaming document upload.
// The client first sends DocMeta, then chunks of file data.
// Data is written to a temporary file to avoid buffering in memory.
func (l *UploadDocumentLogic) UploadDocument(stream rag.Rag_UploadDocumentServer) error {
	// Read the first message (DocMeta)
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	meta := first.GetMeta()
	if meta == nil {
		return errInvalidParam
	}

	// Verify KB exists
	kb, err := l.svcCtx.Storage.GetKB(l.ctx, meta.KbId)
	if err != nil {
		return errKBNotFound
	}
	_ = kb // for future owner validation

	// Enforce max file size
	if meta.FileSize > maxFileSize {
		l.Errorf("file too large: %d bytes (max %d)", meta.FileSize, maxFileSize)
		return fmt.Errorf("file size %d exceeds maximum %d", meta.FileSize, maxFileSize)
	}

	// Create temp file to stage upload data
	tmpFile, err := os.CreateTemp("", tempDirPattern)
	if err != nil {
		l.Errorf("create temp file: %v", err)
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
	}()

	var written int64
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tmpPath)
			return err
		}
		chunk := msg.GetChunk()
		if chunk != nil {
			n, err := tmpFile.Write(chunk)
			if err != nil {
				os.Remove(tmpPath)
				return err
			}
			written += int64(n)
			if written > maxFileSize {
				os.Remove(tmpPath)
				return fmt.Errorf("file size exceeds maximum %d", maxFileSize)
			}
		}
	}
	tmpFile.Close()

	l.Infof("Received %d bytes for document, staged at %s", written, tmpPath)

	// Create document record (use actual written size)
	docID := generateID("doc")
	if err := l.svcCtx.Storage.CreateDocument(l.ctx, docID, meta.KbId,
		meta.FileName, meta.ContentType, written, meta.OssIndex); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Process asynchronously — the goroutine owns tmpPath and cleans it up
	go ProcessAndIndexDocument(context.Background(), l.svcCtx, docID, meta.KbId,
		meta.FileName, meta.ContentType, tmpPath)

	return stream.SendAndClose(&rag.UploadDocResp{
		DocId:  docID,
		Status: "PROCESSING",
	})
}