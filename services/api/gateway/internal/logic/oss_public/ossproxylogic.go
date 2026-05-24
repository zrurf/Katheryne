package oss_public

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type OssProxyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	w      http.ResponseWriter
	r      *http.Request
}

func NewOssProxyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OssProxyLogic {
	return &OssProxyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OssProxyLogic) SetWriter(w http.ResponseWriter) {
	l.w = w
}

func (l *OssProxyLogic) SetRequest(r *http.Request) {
	l.r = r
}

func (l *OssProxyLogic) OssProxy(req *types.OssProxyRequest) error {
	if req.Key == "" {
		l.Errorf("OssProxy: empty key")
		http.Error(l.w, "missing key", http.StatusBadRequest)
		return nil
	}

	l.Infof("OssProxy: requesting download for key=%s, range=%s", req.Key, l.r.Header.Get("Range"))

	// Get file info first for metadata (size, content-type, filename, etag)
	fileInfo, fiErr := l.svcCtx.OssRpc.GetFileInfo(l.ctx, &ossclient.GetFileInfoReq{
		ObjectKey: req.Key,
	})
	var fileName string
	var fileSize int64
	var contentType string
	var etag string
	if fiErr == nil && fileInfo != nil {
		fileName = fileInfo.FileName
		fileSize = fileInfo.Size
		contentType = fileInfo.ContentType
		etag = fileInfo.Etag
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Get pre-signed download URL from OSS service
	dlResp, err := l.svcCtx.OssRpc.GetDownloadURL(l.ctx, &ossclient.GetDownloadURLReq{
		ObjectKey:  req.Key,
		ExpireSecs: 3600,
	})
	if err != nil {
		l.Errorf("GetDownloadURL failed for key=%s: %v", req.Key, err)
		http.Error(l.w, "failed to get download URL", http.StatusInternalServerError)
		return nil
	}

	// Build request to OSS, forwarding Range header if present
	ossReq, err := http.NewRequestWithContext(l.ctx, "GET", dlResp.Url, nil)
	if err != nil {
		l.Errorf("create OSS request failed: %v", err)
		http.Error(l.w, "internal error", http.StatusInternalServerError)
		return nil
	}

	rangeHeader := l.r.Header.Get("Range")
	hasRange := rangeHeader != ""

	if hasRange {
		ossReq.Header.Set("Range", rangeHeader)
	}

	ossResp, err := http.DefaultClient.Do(ossReq)
	if err != nil {
		l.Errorf("fetch file failed for key=%s: %v", req.Key, err)
		http.Error(l.w, "failed to fetch file", http.StatusInternalServerError)
		return nil
	}
	defer ossResp.Body.Close()

	// Common headers
	if fileName != "" {
		l.w.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	}
	l.w.Header().Set("Content-Type", contentType)
	l.w.Header().Set("Accept-Ranges", "bytes")
	l.w.Header().Set("Cache-Control", "public, max-age=3600")
	if etag != "" {
		l.w.Header().Set("ETag", `"`+etag+`"`)
	}

	// Handle range request
	if hasRange {
		if ossResp.StatusCode == http.StatusPartialContent {
			// Forward partial content headers from OSS
			if cr := ossResp.Header.Get("Content-Range"); cr != "" {
				l.w.Header().Set("Content-Range", cr)
			}
			if cl := ossResp.Header.Get("Content-Length"); cl != "" {
				l.w.Header().Set("Content-Length", cl)
			}
			l.w.WriteHeader(http.StatusPartialContent)

			written, _ := io.Copy(l.w, ossResp.Body)
			l.Infof("OssProxy: served %d bytes (range) for key=%s", written, req.Key)
			return nil
		}

		// Range not satisfiable — return 416 with total size
		if fileSize > 0 {
			l.w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		}
		l.w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		l.Infof("OssProxy: range not satisfiable for key=%s", req.Key)
		return nil
	}

	// Full download
	if ossResp.StatusCode != http.StatusOK {
		l.Errorf("OSS returned %d for key=%s", ossResp.StatusCode, req.Key)
		http.Error(l.w, "file not found", http.StatusNotFound)
		return nil
	}

	// Fallback: get file size from OSS response if GetFileInfo didn't provide it
	if fileSize <= 0 {
		if cl := ossResp.Header.Get("Content-Length"); cl != "" {
			if parsed, parseErr := strconv.ParseInt(cl, 10, 64); parseErr == nil {
				fileSize = parsed
			}
		}
	}

	if fileSize > 0 {
		l.w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
	}
	l.w.WriteHeader(http.StatusOK)

	l.Infof("OssProxy: streaming file, content-type=%s, name=%s, size=%d", contentType, fileName, fileSize)

	written, _ := io.Copy(l.w, ossResp.Body)
	l.Infof("OssProxy: served %d bytes for key=%s", written, req.Key)

	return nil
}
