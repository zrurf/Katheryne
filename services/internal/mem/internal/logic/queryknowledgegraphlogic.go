package logic

import (
	"context"

	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryKnowledgeGraphLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryKnowledgeGraphLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryKnowledgeGraphLogic {
	return &QueryKnowledgeGraphLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryKnowledgeGraph traverses the memory knowledge graph from starting entities.
func (l *QueryKnowledgeGraphLogic) QueryKnowledgeGraph(in *mem.QueryKnowledgeGraphReq) (*mem.QueryKnowledgeGraphResp, error) {
	depth := in.Depth
	if depth <= 0 {
		depth = 2
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}

	nodes, edges, err := l.svcCtx.Postgres.QueryGraphTraverse(l.ctx,
		in.TenantId, in.TenantType, in.EntityIds, depth, limit)
	if err != nil {
		logx.Errorf("query graph failed: %v", err)
		return nil, err
	}

	var protoNodes []*mem.GraphNode
	for _, n := range nodes {
		protoNodes = append(protoNodes, &mem.GraphNode{
			EntityId:   n.EntityID,
			Name:       n.Name,
			EntityType: n.EntityType,
		})
	}

	var protoEdges []*mem.GraphEdge
	for _, e := range edges {
		protoEdges = append(protoEdges, &mem.GraphEdge{
			EdgeId:       e.EdgeID,
			Source:       e.Source,
			Target:       e.Target,
			RelationType: e.RelationType,
			Weight:       e.Weight,
		})
	}

	logx.Infof("graph query: entities=%v depth=%d nodes=%d edges=%d",
		in.EntityIds, depth, len(protoNodes), len(protoEdges))

	return &mem.QueryKnowledgeGraphResp{
		Nodes: protoNodes,
		Edges: protoEdges,
	}, nil
}