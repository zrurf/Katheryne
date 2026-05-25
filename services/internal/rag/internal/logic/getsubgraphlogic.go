package logic

import (
	"context"

	"rag/internal/svc"
	"rag/rag/rag"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubGraphLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSubGraphLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubGraphLogic {
	return &GetSubGraphLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSubGraphLogic) GetSubGraph(in *rag.GetSubGraphReq) (*rag.GetSubGraphResp, error) {
	depth := in.Depth
	if depth < 1 {
		depth = 2
	}
	limit := in.Limit
	if limit < 1 || limit > 500 {
		limit = 100
	}

	if l.svcCtx.HugeGraph == nil {
		return &rag.GetSubGraphResp{}, nil
	}

	vertsRaw, edgesRaw, err := l.svcCtx.HugeGraph.GetSubGraph(l.ctx, in.EntityId, int(depth), int(limit))
	if err != nil {
		return nil, err
	}

	var entities []*rag.GraphEntity
	for _, v := range vertsRaw {
		entity := &rag.GraphEntity{}
		if id, ok := v["id"].(string); ok {
			entity.EntityId = id
		}
		if props, ok := v["properties"].(map[string]interface{}); ok {
			if name, ok := props["name"].(string); ok {
				entity.Name = name
			}
			if t, ok := props["type"].(string); ok {
				entity.Type = t
			}
		}
		entities = append(entities, entity)
	}

	var relations []*rag.GraphRelation
	for _, e := range edgesRaw {
		rel := &rag.GraphRelation{}
		if id, ok := e["id"].(string); ok {
			rel.RelationId = id
		}
		if outV, ok := e["outV"].(string); ok {
			rel.SourceEntity = outV
		}
		if inV, ok := e["inV"].(string); ok {
			rel.TargetEntity = inV
		}
		if props, ok := e["properties"].(map[string]interface{}); ok {
			if rt, ok := props["relation_type"].(string); ok {
				rel.RelationType = rt
			}
			if chunkID, ok := props["chunk_id"].(string); ok {
				rel.ChunkId = chunkID
			}
		}
		relations = append(relations, rel)
	}

	return &rag.GetSubGraphResp{
		Entities:  entities,
		Relations: relations,
	}, nil
}