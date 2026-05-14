package oss

import (
	"context"
	"io"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      io.Reader
}

func NewUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadLogic) SetBody(r io.Reader) {
	l.r = r
}

func (l *UploadLogic) Upload() (resp *types.UploadResponse, err error) {
	stream, err := l.svcCtx.OssRpc.UploadPart(l.ctx)
	if err != nil {
		l.Errorf("Upload stream creation failed: %v", err)
		return nil, err
	}

	buf := make([]byte, 32*1024)
	for {
		n, readErr := l.r.Read(buf)
		if n > 0 {
			sendErr := stream.Send(&ossclient.UploadPartReq{
				UploadId:   "simple",
				PartNumber: 1,
				Data:       buf[:n],
			})
			if sendErr != nil {
				l.Errorf("Upload send failed: %v", sendErr)
				return nil, sendErr
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			l.Errorf("Upload read failed: %v", readErr)
			return nil, readErr
		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		l.Errorf("Upload close failed: %v", err)
		return nil, err
	}
	return &types.UploadResponse{}, nil
}