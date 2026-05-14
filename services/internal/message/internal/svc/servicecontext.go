package svc

import (
	"context"
	"fmt"
	"net/url"

	"message/internal/config"
	"message/internal/dao"
	"user/userclient"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	DbPool     *pgxpool.Pool
	Redis      *redis.Client
	MessageDao *dao.MessageDao
	RedisDao   *dao.RedisDao
	UserRpc    userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.MessageDB.Host == "" || c.MessageDB.User == "" || c.MessageDB.DBName == "" {
		panic(fmt.Sprintf("database config is incomplete: host=%q user=%q dbname=%q", c.MessageDB.Host, c.MessageDB.User, c.MessageDB.DBName))
	}

	connString := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.MessageDB.User, c.MessageDB.Password),
		Host:     fmt.Sprintf("%s:%d", c.MessageDB.Host, c.MessageDB.Port),
		Path:     c.MessageDB.DBName,
		RawQuery: "sslmode=disable",
	}).String()

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	msgDao := dao.NewMessageDao(dbPool)
	redisDao := dao.NewRedisDao(rdb)

	return &ServiceContext{
		Config:     c,
		DbPool:     dbPool,
		Redis:      rdb,
		MessageDao: msgDao,
		RedisDao:   redisDao,
		UserRpc:    userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
