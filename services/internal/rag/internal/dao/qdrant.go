package dao

import (
	"context"
	"fmt"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QdrantDao struct {
	client     pb.PointsClient
	collections pb.CollectionsClient
	conn       *grpc.ClientConn
	vectorDim  uint64
}

func NewQdrantDao(host string, port int, useTLS bool, apiKey string, vectorDim int) (*QdrantDao, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if apiKey != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&apiKeyCred{key: apiKey}))
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

func (q *QdrantDao) Close() error {
	return q.conn.Close()
}

// EnsureCollection creates a collection for a knowledge base if not exists.
func (q *QdrantDao) EnsureCollection(ctx context.Context, kbID string) error {
	exists, err := q.collectionExists(ctx, kbID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = q.collections.Create(ctx, &pb.CreateCollection{
		CollectionName: kbID,
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
		return fmt.Errorf("create collection %s: %w", kbID, err)
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

// DeleteCollection removes a whole knowledge base collection.
func (q *QdrantDao) DeleteCollection(ctx context.Context, kbID string) error {
	_, err := q.collections.Delete(ctx, &pb.DeleteCollection{
		CollectionName: kbID,
	})
	return err
}

// UpsertPoints inserts or updates vector points (chunks).
func (q *QdrantDao) UpsertPoints(ctx context.Context, kbID string, points []*pb.PointStruct) error {
	_, err := q.client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: kbID,
		Points:         points,
	})
	return err
}

// Search performs vector similarity search.
func (q *QdrantDao) Search(ctx context.Context, kbID string, vector []float32, topK uint64, filter *pb.Filter) ([]*pb.ScoredPoint, error) {
	resp, err := q.client.Search(ctx, &pb.SearchPoints{
		CollectionName: kbID,
		Vector:         vector,
		Limit:          topK,
		Filter:         filter,
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// DeletePoints removes specific chunk points.
func (q *QdrantDao) DeletePoints(ctx context.Context, kbID string, pointIDs []string) error {
	ids := make([]*pb.PointId, len(pointIDs))
	for i, id := range pointIDs {
		ids[i] = &pb.PointId{PointIdOptions: &pb.PointId_Uuid{Uuid: id}}
	}
	_, err := q.client.Delete(ctx, &pb.DeletePoints{
		CollectionName: kbID,
		Points:         &pb.PointsSelector{PointsSelectorOneOf: &pb.PointsSelector_Points{Points: &pb.PointsIdsList{Ids: ids}}},
	})
	return err
}

// Scroll retrieves all points in a collection (with pagination).
func (q *QdrantDao) Scroll(ctx context.Context, kbID string, offset *pb.PointId, limit uint32) (*pb.ScrollResponse, error) {
	return q.client.Scroll(ctx, &pb.ScrollPoints{
		CollectionName: kbID,
		Offset:         offset,
		Limit:          &limit,
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
}

type apiKeyCred struct {
	key string
}

func (c *apiKeyCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"api-key": c.key}, nil
}

func (c *apiKeyCred) RequireTransportSecurity() bool {
	return false
}