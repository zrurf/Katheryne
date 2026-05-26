package logic

import (
	"context"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearMemoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearMemoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearMemoriesLogic {
	return &ClearMemoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearMemoriesLogic) ClearMemories(in *mem.ClearMemoriesReq) (*mem.ClearMemoriesResp, error) {
	count, err := l.svcCtx.Postgres.ClearMemories(l.ctx, in.TenantId, in.TenantType)
	if err != nil {
		logx.Errorf("clear memories failed: %v", err)
		return nil, err
	}
	_ = l.svcCtx.Qdrant.DeleteTenantPoints(l.ctx, in.TenantId)

	logx.Infof("cleared %d memories for tenant=%s", count, in.TenantId)
	return &mem.ClearMemoriesResp{DeletedCount: count}, nil
}