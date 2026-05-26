package svc

import (
	"fmt"

	"mem/internal/config"
	"mem/internal/dao"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config     config.Config
	Postgres   *dao.PostgresDao
	Qdrant     *dao.QdrantDao
}

func NewServiceContext(c config.Config) *ServiceContext {
	pgDao, err := dao.NewPostgresDao(c.MemDB.Host, c.MemDB.Port, c.MemDB.User, c.MemDB.Password, c.MemDB.DBName)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to postgres: %v", err))
	}

	qdrantDao, err := dao.NewQdrantDao(c.Qdrant.Host, c.Qdrant.Port, c.Qdrant.UseTLS, c.Qdrant.APIKey, c.Qdrant.VectorDim)
	if err != nil {
		pgDao.Close()
		panic(fmt.Sprintf("failed to connect to qdrant: %v", err))
	}

	ctx := &ServiceContext{
		Config:   c,
		Postgres: pgDao,
		Qdrant:   qdrantDao,
	}

	logx.Info("mem service context initialized")
	return ctx
}