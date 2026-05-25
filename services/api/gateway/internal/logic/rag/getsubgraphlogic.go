package rag

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"
	"rag/ragclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubGraphLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubGraphLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubGraphLogic {
	return &GetSubGraphLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubGraphLogic) GetSubGraph(req *types.GetSubGraphRequest) (resp *types.GetSubGraphResponse, err error) {
	result, err := l.svcCtx.RagRpc.GetSubGraph(l.ctx, &ragclient.GetSubGraphReq{
		KbId:     req.KbID,
		EntityId: req.EntityID,
		Depth:    int32(req.Depth),
		Limit:    int32(req.Limit),
	})
	if err != nil {
		return nil, err
	}
	entities := make([]types.GraphEntityInfo, 0, len(result.Entities))
	for _, e := range result.Entities {
		props := make(map[string]string)
		for _, p := range e.Properties {
			props[p] = ""
		}
		entities = append(entities, types.GraphEntityInfo{
			EntityID:   e.EntityId,
			Name:       e.Name,
			Type:       e.Type,
			Properties: props,
		})
	}
	relations := make([]types.GraphRelationInfo, 0, len(result.Relations))
	for _, r := range result.Relations {
		relations = append(relations, types.GraphRelationInfo{
			RelationID:   r.RelationId,
			SourceEntity: r.SourceEntity,
			TargetEntity: r.TargetEntity,
			RelationType: r.RelationType,
			ChunkID:      r.ChunkId,
		})
	}
	return &types.GetSubGraphResponse{Entities: entities, Relations: relations}, nil
}
