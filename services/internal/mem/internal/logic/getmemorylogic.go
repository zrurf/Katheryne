package logic

import (
	"context"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMemoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemoryLogic {
	return &GetMemoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMemoryLogic) GetMemory(in *mem.GetMemoryReq) (*mem.GetMemoryResp, error) {
	row, err := l.svcCtx.Postgres.GetMemory(l.ctx, in.MemoryId)
	if err != nil {
		logx.Errorf("get memory failed: %v", err)
		return nil, err
	}
	if row == nil {
		return &mem.GetMemoryResp{}, nil
	}

	// Record access for importance decay/boost
	_ = l.svcCtx.Postgres.RecordAccess(l.ctx, in.MemoryId)

	return &mem.GetMemoryResp{Memory: toProto(row)}, nil
}