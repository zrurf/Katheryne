package oss

import (
	"context"
	"io"

	"gateway/internal/svc"
	"gateway/internal/types"
	"oss/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadPartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      io.Reader
}

func NewUploadPartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPartLogic {
	return &UploadPartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadPartLogic) SetBody(r io.Reader) {
	l.r = r
}

func (l *UploadPartLogic) UploadPart(req *types.UploadPartRequest) (resp *types.UploadPartResponse, err error) {
	stream, err := l.svcCtx.OssRpc.UploadPart(l.ctx)
	if err != nil {
		l.Errorf("UploadPart stream creation failed: %v", err)
		return nil, err
	}

	buf := make([]byte, 32*1024)
	for {
		n, readErr := l.r.Read(buf)
		if n > 0 {
			sendErr := stream.Send(&ossclient.UploadPartReq{
				UploadId:   req.UploadID,
				PartNumber: int32(req.PartNumber),
				Data:       buf[:n],
			})
			if sendErr != nil {
				l.Errorf("UploadPart send failed: %v", sendErr)
				return nil, sendErr
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			l.Errorf("UploadPart read failed: %v", readErr)
			return nil, readErr
		}
	}

	result, err := stream.CloseAndRecv()
	if err != nil {
		l.Errorf("UploadPart close failed: %v", err)
		return nil, err
	}
	return &types.UploadPartResponse{
		ETag:       result.Etag,
		PartNumber: int(result.PartNumber),
	}, nil
}
