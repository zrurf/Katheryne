package oss_public

import (
	"context"
	"io"
	"net/http"

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

	l.Infof("OssProxy: requesting download URL for key=%s", req.Key)

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

	// Get file info for proper filename and content-type
	fileInfo, fiErr := l.svcCtx.OssRpc.GetFileInfo(l.ctx, &ossclient.GetFileInfoReq{
		ObjectKey: req.Key,
	})
	var fileName string
	if fiErr == nil && fileInfo != nil {
		fileName = fileInfo.FileName
	}

	l.Infof("OssProxy: got presigned URL, fetching from OSS")

	// Fetch file content from OSS
	httpResp, err := http.Get(dlResp.Url)
	if err != nil {
		l.Errorf("fetch file failed for key=%s: %v", req.Key, err)
		http.Error(l.w, "failed to fetch file", http.StatusInternalServerError)
		return nil
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		l.Errorf("OSS returned %d for key=%s", httpResp.StatusCode, req.Key)
		http.Error(l.w, "file not found", http.StatusNotFound)
		return nil
	}

	l.Infof("OssProxy: streaming file, content-type=%s, name=%s", httpResp.Header.Get("Content-Type"), fileName)

	// Stream file content to client
	contentType := httpResp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	l.w.Header().Set("Content-Type", contentType)
	l.w.Header().Set("Cache-Control", "public, max-age=3600")
	if fileName != "" {
		l.w.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	}
	l.w.WriteHeader(http.StatusOK)

	written, _ := io.Copy(l.w, httpResp.Body)
	l.Infof("OssProxy: served %d bytes for key=%s", written, req.Key)
	return nil
}
