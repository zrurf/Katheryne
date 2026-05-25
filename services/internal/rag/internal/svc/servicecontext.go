package svc

import (
	"context"
	"fmt"

	"rag/internal/config"
	"rag/internal/dao"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config    config.Config
	Redis     *redis.Client
	RagDB     *pgxpool.Pool
	Storage   *dao.StorageDao
	Qdrant    *dao.QdrantDao
	HugeGraph *dao.HugeGraphDao
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	// PostgreSQL for metadata
	if c.RagDB.Host == "" || c.RagDB.User == "" || c.RagDB.DBName == "" {
		panic(fmt.Sprintf("rag database config is incomplete: host=%q user=%q dbname=%q",
			c.RagDB.Host, c.RagDB.User, c.RagDB.DBName))
	}

	ragDB, err := pgxpool.New(context.Background(),
		dao.BuildDSN(c.RagDB.Host, c.RagDB.User, c.RagDB.Password, c.RagDB.DBName, c.RagDB.Port))
	if err != nil {
		panic(err)
	}

	storage := dao.NewStorageDao(ragDB)

	// Qdrant vector database
	qdrant, err := dao.NewQdrantDao(c.Qdrant.Host, c.Qdrant.Port, c.Qdrant.UseTLS, c.Qdrant.APIKey, c.Qdrant.VectorDim)
	if err != nil {
		logx.Errorf("Qdrant init warning (service will start anyway): %v", err)
	}

	// HugeGraph
	hugeGraph := dao.NewHugeGraphDao(c.HugeGraph.BaseURL, c.HugeGraph.Graph)
	// Ensure schema (non-fatal on startup)
	if err := hugeGraph.EnsureSchema(context.Background()); err != nil {
		logx.Errorf("HugeGraph schema init warning: %v", err)
	}

	return &ServiceContext{
		Config:    c,
		Redis:     rdb,
		RagDB:     ragDB,
		Storage:   storage,
		Qdrant:    qdrant,
		HugeGraph: hugeGraph,
	}
}