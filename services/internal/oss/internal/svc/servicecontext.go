package svc

import (
	"oss/internal/config"
	"oss/internal/dao"

	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config    config.Config
	Redis     *redis.Client
	Storage   *dao.StorageDao
	RedisDao  *dao.RedisDao
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
		panic(err)
	}

	redisDao := dao.NewRedisDao(rdb)

	return &ServiceContext{
		Config:    c,
		Redis:     rdb,
		Storage:   storage,
		RedisDao:  redisDao,
	}
}
