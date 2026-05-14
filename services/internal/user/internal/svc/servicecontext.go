package svc

import (
	"context"
	"fmt"
	"net/url"
	"user/internal/config"
	"user/internal/dao"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config config.Config

	DbPool   *pgxpool.Pool
	Redis    *redis.Client
	UserDao  *dao.UserDao
	RedisDao *dao.RedisDao
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.UserDB.Host == "" || c.UserDB.User == "" || c.UserDB.DBName == "" {
		panic(fmt.Sprintf("database config is incomplete: host=%q user=%q dbname=%q", c.UserDB.Host, c.UserDB.User, c.UserDB.DBName))
	}

	connString := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.UserDB.User, c.UserDB.Password),
		Host:     fmt.Sprintf("%s:%d", c.UserDB.Host, c.UserDB.Port),
		Path:     c.UserDB.DBName,
		RawQuery: "sslmode=disable",
	}).String()

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	userDao := dao.NewUserDao(dbPool)
	redisDao := dao.NewRedisDao(rdb)

	return &ServiceContext{
		Config:   c,
		DbPool:   dbPool,
		Redis:    rdb,
		UserDao:  userDao,
		RedisDao: redisDao,
	}
}
