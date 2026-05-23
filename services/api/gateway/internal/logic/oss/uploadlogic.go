package oss

import (
	"context"
	"io"
	"net/url"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	file        io.Reader
	fileName    string
	contentType string
	host        string
}

func NewUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadLogic) SetFile(r io.Reader) {
	l.file = r
}

func (l *UploadLogic) SetFileName(name string) {
	if name == "" {
		name = "upload"
	}
	l.fileName = name
}

func (l *UploadLogic) SetContentType(ct string) {
	if ct == "" {
		ct = "application/octet-stream"
	}
	l.contentType = ct
}

func (l *UploadLogic) SetHost(host string) {
	l.host = host
}

func (l *UploadLogic) Upload() (resp *types.UploadResponse, err error) {
	if l.file == nil {
		return nil, io.ErrUnexpectedEOF
	}

	stream, err := l.svcCtx.OssRpc.SimpleUpload(l.ctx)
	if err != nil {
		l.Errorf("SimpleUpload stream creation failed: %v", err)
		return nil, err
	}

	if err := stream.Send(&ossclient.SimpleUploadReq{
		Data: &ossclient.SimpleUploadReq_Meta{
			Meta: &ossclient.FileMeta{
				FileName:    l.fileName,
				ContentType: l.contentType,
			},
		},
	}); err != nil {
		l.Errorf("SimpleUpload meta send failed: %v", err)
		return nil, err
	}

	buf := make([]byte, 32*1024)
	for {
		n, readErr := l.file.Read(buf)
		if n > 0 {
			sendErr := stream.Send(&ossclient.SimpleUploadReq{
				Data: &ossclient.SimpleUploadReq_Chunk{
					Chunk: buf[:n],
				},
			})
			if sendErr != nil {
				l.Errorf("SimpleUpload chunk send failed: %v", sendErr)
				return nil, sendErr
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			l.Errorf("SimpleUpload read failed: %v", readErr)
			return nil, readErr
		}
	}

	result, err := stream.CloseAndRecv()
	if err != nil {
		l.Errorf("SimpleUpload close failed: %v", err)
		return nil, err
	}

	// Return only the path (no scheme/host) so the App can assemble the
	// full URL using its own server host. This avoids issues with Docker-
	// internal hostnames (e.g. rustfs:9000) leaking to clients.
	proxyPath := "/api/v1/oss/file?" + url.Values{"key": {result.OssIndex}}.Encode()

	return &types.UploadResponse{
		Url:      proxyPath,
		OssIndex: result.OssIndex,
		IndexId:  result.IndexId,
		FileName: result.FileName,
		Size:     result.Size,
	}, nil
}
