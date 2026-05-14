package svc

import (
	"context"
	"fmt"
	"net/url"

	"social/internal/config"
	"social/internal/dao"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config config.Config

	DbPool     *pgxpool.Pool
	UserDbPool *pgxpool.Pool
	Redis      *redis.Client
	SocialDao  *dao.SocialDao
	UserDBDao  *dao.UserDBDao
	RedisDao   *dao.RedisDao
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

	var userDbPool *pgxpool.Pool
	if c.UserDB.Host != "" {
		userConnString := (&url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(c.UserDB.User, c.UserDB.Password),
			Host:     fmt.Sprintf("%s:%d", c.UserDB.Host, c.UserDB.Port),
			Path:     c.UserDB.DBName,
			RawQuery: "sslmode=disable",
		}).String()
		userDbPool, err = pgxpool.New(context.Background(), userConnString)
		if err != nil {
			panic(err)
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	socialDao := dao.NewSocialDao(dbPool)
	userDBDao := dao.NewUserDBDao(userDbPool)
	redisDao := dao.NewRedisDao(rdb)

	return &ServiceContext{
		Config:     c,
		DbPool:     dbPool,
		UserDbPool: userDbPool,
		Redis:      rdb,
		SocialDao:  socialDao,
		UserDBDao:  userDBDao,
		RedisDao:   redisDao,
	}
}
