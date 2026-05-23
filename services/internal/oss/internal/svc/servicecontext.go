package svc

import (
	"oss/internal/config"
	"oss/internal/dao"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config   config.Config
	Redis    *redis.Client
	Storage  *dao.StorageDao
	RedisDao *dao.RedisDao
	Hashers  sync.Map // uploadID → *blake3.Hasher (for streaming hash during multipart upload)
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	storage, err := dao.NewStorageDao(
		c.RustFS.Endpoint,
		c.RustFS.AccessKey,
		c.RustFS.SecretKey,
		c.RustFS.Bucket,
		c.RustFS.Region,
		c.RustFS.UseSSL,
	)
	if err != nil {
		// Non-fatal: service can start and retry on first request.
		// Common when MinIO is not yet ready or Docker DNS is temporarily unavailable.
		logx.Errorf("StorageDao init warning (service will start anyway): %v", err)
	}

	redisDao := dao.NewRedisDao(rdb)

	return &ServiceContext{
		Config:   c,
		Redis:    rdb,
		Storage:  storage,
		RedisDao: redisDao,
	}
}
