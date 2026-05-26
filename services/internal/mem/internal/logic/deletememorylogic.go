package logic

import (
	"context"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMemoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMemoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMemoryLogic {
	return &DeleteMemoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteMemoryLogic) DeleteMemory(in *mem.DeleteMemoryReq) (*mem.DeleteMemoryResp, error) {
	if err := l.svcCtx.Postgres.DeleteMemory(l.ctx, in.MemoryId); err != nil {
		logx.Errorf("delete memory failed: %v", err)
		return nil, err
	}
	_ = l.svcCtx.Qdrant.DeletePoint(l.ctx, in.MemoryId)
	logx.Infof("memory deleted: id=%s", in.MemoryId)
	return &mem.DeleteMemoryResp{}, nil
}