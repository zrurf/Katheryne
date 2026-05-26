package logic

import (
	"context"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMemoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMemoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMemoryLogic {
	return &UpdateMemoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateMemoryLogic) UpdateMemory(in *mem.UpdateMemoryReq) (*mem.UpdateMemoryResp, error) {
	metadata := in.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	if err := l.svcCtx.Postgres.UpdateMemory(l.ctx, in.MemoryId, in.Content, in.Importance, metadata); err != nil {
		logx.Errorf("update memory failed: %v", err)
		return nil, err
	}

	logx.Infof("memory updated: id=%s", in.MemoryId)
	return &mem.UpdateMemoryResp{}, nil
}