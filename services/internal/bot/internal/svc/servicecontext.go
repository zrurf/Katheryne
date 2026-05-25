package svc

import (
	"context"
	"fmt"
	"net/url"

	"bot/internal/config"
	"bot/internal/dao"
	"conversation/conversationclient"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config
	Redis           *redis.Client
	BotDB           *pgxpool.Pool
	SocialDB        *pgxpool.Pool
	ConversationRpc conversationclient.Conversation
	BotDao          *dao.BotDao
	InstDao         *dao.InstallationDao
	OAuthDao        *dao.OAuthDao
	EventDao        *dao.EventDao
	TemplateDao     *dao.TemplateDao
	InstanceDao     *dao.InstanceDao
}

func buildDSN(cfg config.DBConfig) string {
	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     cfg.DBName,
		RawQuery: "sslmode=disable",
	}).String()
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.BotDB.Host == "" || c.BotDB.User == "" || c.BotDB.DBName == "" {
		panic(fmt.Sprintf("database config is incomplete: host=%q user=%q dbname=%q", c.BotDB.Host, c.BotDB.User, c.BotDB.DBName))
	}

	botDB, err := pgxpool.New(context.Background(), buildDSN(c.BotDB))
	if err != nil {
		panic(err)
	}

	var socialDB *pgxpool.Pool
	if c.SocialDB.Host != "" && c.SocialDB.DBName != "" {
		socialDB, err = pgxpool.New(context.Background(), buildDSN(c.SocialDB))
		if err != nil {
			panic(err)
		}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	conversationRpc := conversationclient.NewConversation(zrpc.MustNewClient(c.ConversationRpc))

	botDao := dao.NewBotDao(botDB)
	instDao := dao.NewInstallationDao(botDB, socialDB)
	oauthDao := dao.NewOAuthDao(botDB, redisClient)
	eventDao := dao.NewEventDao(botDB)
	templateDao := dao.NewTemplateDao(botDB)
	instanceDao := dao.NewInstanceDao(botDB, templateDao)

	return &ServiceContext{
		Config:          c,
		Redis:           redisClient,
		BotDB:           botDB,
		SocialDB:        socialDB,
		ConversationRpc: conversationRpc,
		BotDao:          botDao,
		InstDao:         instDao,
		OAuthDao:        oauthDao,
		EventDao:        eventDao,
		TemplateDao:     templateDao,
		InstanceDao:     instanceDao,
	}
}
