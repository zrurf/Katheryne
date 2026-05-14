package svc

import (
	"auth/internal/config"
	"auth/internal/dao"
	"auth/internal/module"
	"context"
	"crypto"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/bytemare/ksf"
	"github.com/bytemare/opaque"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sony/sonyflake"
)

type ServiceContext struct {
	Config config.Config

	DbPool *pgxpool.Pool
	Redis  *redis.Client

	OpaqueSvc  *module.OpaqueService
	UserDao    *dao.UserDao
	SessionDao *dao.SessionDao

	SonyFlake *sonyflake.Sonyflake
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

	redis := redis.NewClient(&redis.Options{
		Addr: c.RedisAddr,
	})

	oprfSeed, err := os.ReadFile(c.Opaque.OPRFSeedFile)
	if err != nil {
		panic("cannot read oprf seed file")
	}
	pubKey, err := os.ReadFile(c.Opaque.ServerPublicKeyFile)
	if err != nil {
		panic("cannot read public key file")
	}
	priKey, err := os.ReadFile(c.Opaque.ServerSecretKeyFile)
	if err != nil {
		panic("cannot read private key file")
	}

	opaqueSvc, err := module.NewOpaqueService(&module.OpaqueConfig{
		Config: &opaque.Configuration{
			OPRF:    opaque.RistrettoSha512,
			KDF:     crypto.SHA512,
			MAC:     crypto.SHA512,
			Hash:    crypto.SHA512,
			KSF:     ksf.Argon2id,
			AKE:     opaque.RistrettoSha512,
			Context: nil,
		},
		OprfSeed:        oprfSeed,
		ServerPublicKey: pubKey,
		ServerSecretKey: priKey,
	})

	if err != nil {
		panic(err)
	}

	userDao := dao.NewUserDao(dbPool)
	sessionDao := dao.NewSessionDao(redis)

	startTime, err := time.Parse(time.RFC3339, c.Snoyflake.StartTime)
	if err != nil {
		startTime = time.UnixMilli(0)
	}

	sf := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: startTime,
		MachineID: func() (uint16, error) {
			return c.Snoyflake.MachineID, nil
		},
	})

	return &ServiceContext{
		Config:     c,
		DbPool:     dbPool,
		Redis:      redis,
		OpaqueSvc:  opaqueSvc,
		UserDao:    userDao,
		SessionDao: sessionDao,
		SonyFlake:  sf,
	}
}
