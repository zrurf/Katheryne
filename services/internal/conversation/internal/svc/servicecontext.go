package svc

import (
	"context"
	"conversation/internal/config"
	"conversation/internal/dao"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config config.Config

	DbPool          *pgxpool.Pool
	Redis           *redis.Client
	ConversationDao *dao.ConversationDao
	RedisDao        *dao.RedisDao
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.SocialDB.Host == "" || c.SocialDB.User == "" || c.SocialDB.DBName == "" {
		panic(fmt.Sprintf("database config is incomplete: host=%q user=%q dbname=%q", c.SocialDB.Host, c.SocialDB.User, c.SocialDB.DBName))
	}

	connString := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.SocialDB.User, c.SocialDB.Password),
		Host:     fmt.Sprintf("%s:%d", c.SocialDB.Host, c.SocialDB.Port),
		Path:     c.SocialDB.DBName,
		RawQuery: "sslmode=disable",
	}).String()

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	convDao := dao.NewConversationDao(dbPool)
	redisDao := dao.NewRedisDao(rdb)

	return &ServiceContext{
		Config:          c,
		DbPool:          dbPool,
		Redis:           rdb,
		ConversationDao: convDao,
		RedisDao:        redisDao,
	}
}
