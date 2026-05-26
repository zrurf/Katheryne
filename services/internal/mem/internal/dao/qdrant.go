package dao

import (
	"context"
	"fmt"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const memoryCollectionName = "katheryne_memories"

// QdrantDao wraps Qdrant gRPC client for memory vector operations.
type QdrantDao struct {
	client      pb.PointsClient
	collections pb.CollectionsClient
	conn        *grpc.ClientConn
	vectorDim   uint64
}

// NewQdrantDao creates a new Qdrant DAO.
func NewQdrantDao(host string, port int, useTLS bool, apiKey string, vectorDim int) (*QdrantDao, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("qdrant connect failed: %w", err)
	}

	return &QdrantDao{
		client:      pb.NewPointsClient(conn),
		collections: pb.NewCollectionsClient(conn),
		conn:        conn,
		vectorDim:   uint64(vectorDim),
	}, nil
}

// Close closes the gRPC connection.
func (q *QdrantDao) Close() error {
	return q.conn.Close()
}

// EnsureCollection creates the memory collection if it doesn't exist.
func (q *QdrantDao) EnsureCollection(ctx context.Context) error {
	exists, err := q.collectionExists(ctx, memoryCollectionName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = q.collections.Create(ctx, &pb.CreateCollection{
		CollectionName: memoryCollectionName,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     q.vectorDim,
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create collection %s: %w", memoryCollectionName, err)
	}
	return nil
}

func (q *QdrantDao) collectionExists(ctx context.Context, name string) (bool, error) {
	resp, err := q.collections.List(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		return false, err
	}
	for _, c := range resp.Collections {
		if c.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// UpsertPoint inserts or updates a vector point in the memory collection.
func (q *QdrantDao) UpsertPoint(ctx context.Context, pointID string, vector []float32, tenantID, tenantType string) error {
	_, err := q.client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: memoryCollectionName,
		Points: []*pb.PointStruct{
			{
				Id:      pb.NewID(pointID),
				Vectors: pb.NewVectors(vector...),
				Payload: map[string]*pb.Value{
					"tenant_id":   {Kind: &pb.Value_StringValue{StringValue: tenantID}},
					"tenant_type": {Kind: &pb.Value_StringValue{StringValue: tenantType}},
				},
			},
		},
	})
	return err
}

// DeletePoint removes a vector point.
func (q *QdrantDao) DeletePoint(ctx context.Context, pointID string) error {
	_, err := q.client.Delete(ctx, &pb.DeletePoints{
		CollectionName: memoryCollectionName,
		Points:         pb.NewPointsSelector(pb.NewID(pointID)),
	})
	return err
}

// DeleteTenantPoints removes all points for a tenant.
func (q *QdrantDao) DeleteTenantPoints(ctx context.Context, tenantID string) error {
	_, err := q.client.Delete(ctx, &pb.DeletePoints{
		CollectionName: memoryCollectionName,
		Points: pb.NewPointsSelectorFilter(&pb.Filter{
			Must: []*pb.Condition{
				pb.NewMatchKeyword("tenant_id", tenantID),
			},
		}),
	})
	return err
}

// SearchVector searches for nearest neighbor vectors.
// Returns point IDs with similarity scores.
func (q *QdrantDao) SearchVector(ctx context.Context, vector []float32, tenantID string, topK uint64) ([]VectorHit, error) {
	filter := &pb.Filter{
		Must: []*pb.Condition{
			pb.NewMatchKeyword("tenant_id", tenantID),
		},
	}

	resp, err := q.client.Search(ctx, &pb.SearchPoints{
		CollectionName: memoryCollectionName,
		Vector:         vector,
		Limit:          topK,
		Filter:         filter,
		WithPayload:    pb.NewWithPayload(false),
	})
	if err != nil {
		return nil, err
	}

	var hits []VectorHit
	for _, scored := range resp.Result {
		hits = append(hits, VectorHit{
			PointID: scored.Id.GetUuid(),
			Score:   float64(scored.Score),
		})
	}
	return hits, nil
}

// VectorHit represents a single vector search result.
type VectorHit struct {
	PointID string
	Score   float64
}
