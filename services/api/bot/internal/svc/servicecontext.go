package svc

import (
	"context"
	"fmt"
	"net/url"

	"bot/internal/config"
	"bot/internal/dao"
	"bot/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config            config.Config
	AuthMiddleware    rest.Middleware
	BotAuthMiddleware rest.Middleware
	Redis             *redis.Client
	DB                *pgxpool.Pool
	WebhookDeliverer  *WebhookDeliverer
	BotDao            *dao.BotDao
	InstallationDao   *dao.InstallationDao
	OAuthDao          *dao.OAuthDao
	EventDao          *dao.EventDao
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.BotDB.Host == "" || c.BotDB.User == "" || c.BotDB.DBName == "" {
		panic(fmt.Sprintf("database config is incomplete: host=%q user=%q dbname=%q", c.BotDB.Host, c.BotDB.User, c.BotDB.DBName))
	}

	connString := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.BotDB.User, c.BotDB.Password),
		Host:     fmt.Sprintf("%s:%d", c.BotDB.Host, c.BotDB.Port),
		Path:     c.BotDB.DBName,
		RawQuery: "sslmode=disable",
	}).String()

	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	botDao := dao.NewBotDao(db)
	installationDao := dao.NewInstallationDao(db)
	oauthDao := dao.NewOAuthDao(db, redisClient)
	eventDao := dao.NewEventDao(db)

	webhookDeliverer := NewWebhookDeliverer(eventDao)
	webhookDeliverer.Start()

	return &ServiceContext{
		Config:            c,
		AuthMiddleware:    middleware.NewAuthMiddleware(redisClient).Handle,
		BotAuthMiddleware: middleware.NewBotAuthMiddleware(redisClient).Handle,
		Redis:             redisClient,
		DB:                db,
		WebhookDeliverer:  webhookDeliverer,
		BotDao:            botDao,
		InstallationDao:   installationDao,
		OAuthDao:          oauthDao,
		EventDao:          eventDao,
	}
}
