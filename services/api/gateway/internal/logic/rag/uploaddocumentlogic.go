package rag

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"gateway/internal/svc"
	"gateway/internal/types"
	ossclient "oss/ossclient"
	ragpb "rag/rag/rag"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadDocumentLogic {
	return &UploadDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadDocumentLogic) UploadDocument(req *types.UploadDocRequest) (resp *types.UploadDocResponse, err error) {
	// Step 1: Get download URL from OSS service
	dlResp, err := l.svcCtx.OssRpc.GetDownloadURL(l.ctx, &ossclient.GetDownloadURLReq{
		ObjectKey:  req.OssKey,
		ExpireSecs: 3600,
	})
	if err != nil {
		l.Errorf("GetDownloadURL failed for key=%s: %v", req.OssKey, err)
		return nil, err
	}

	// Step 2: Download file content from the pre-signed URL
	httpResp, err := downloadFromURL(l.ctx, dlResp.Url)
	if err != nil {
		l.Errorf("downloadFromURL failed for key=%s: %v", req.OssKey, err)
		return nil, err
	}
	defer httpResp.Close()
	fileData, err := io.ReadAll(httpResp)
	if err != nil {
		l.Errorf("read file data failed: %v", err)
		return nil, err
	}
	l.Infof("Downloaded file from OSS: key=%s, size=%d", req.OssKey, len(fileData))

	// Step 3: Open client-streaming gRPC to RAG service
	stream, err := l.svcCtx.RagRpc.UploadDocument(l.ctx)
	if err != nil {
		return nil, err
	}

	// Step 4: Send metadata (DocMeta) as the first message
	if err := stream.Send(&ragclient.UploadDocReq{
		Data: &ragpb.UploadDocReq_Meta{
			Meta: &ragclient.DocMeta{
				KbId:        req.KbID,
				OssIndex:    req.OssKey,
				FileName:    req.FileName,
				ContentType: req.ContentType,
				FileSize:    int64(len(fileData)),
			},
		},
	}); err != nil {
		return nil, err
	}

	// Step 5: Stream file data in chunks
	chunkSize := 64 * 1024 // 64KB chunks
	for i := 0; i < len(fileData); i += chunkSize {
		end := i + chunkSize
		if end > len(fileData) {
			end = len(fileData)
		}
		if err := stream.Send(&ragclient.UploadDocReq{
			Data: &ragpb.UploadDocReq_Chunk{
				Chunk: fileData[i:end],
			},
		}); err != nil {
			return nil, err
		}
	}

	// Step 6: Close stream and receive response
	result, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	l.Infof("Document upload completed: doc_id=%s, status=%s", result.DocId, result.Status)
	return &types.UploadDocResponse{
		DocID:  result.DocId,
		Status: result.Status,
	}, nil
}

// downloadFromURL downloads file content from a pre-signed URL.
func downloadFromURL(ctx context.Context, urlStr string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("OSS download returned status %d", resp.StatusCode)
	}
	return resp.Body, nil
}
