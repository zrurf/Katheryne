package logic

import (
	"context"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMemoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMemoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMemoriesLogic {
	return &ListMemoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListMemoriesLogic) ListMemories(in *mem.ListMemoriesReq) (*mem.ListMemoriesResp, error) {
	rows, total, err := l.svcCtx.Postgres.ListMemories(l.ctx, in.TenantId, in.TenantType, int32(in.Type), in.Page, in.Size)
	if err != nil {
		logx.Errorf("list memories failed: %v", err)
		return nil, err
	}

	var list []*mem.MemoryItem
	for _, row := range rows {
		list = append(list, toProto(row))
	}

	return &mem.ListMemoriesResp{List: list, Total: total}, nil
}