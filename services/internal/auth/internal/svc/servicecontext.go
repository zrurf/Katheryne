package svc

import (
	"auth/internal/config"
	"auth/internal/dao"
	"auth/internal/module"
	"context"
	"crypto"
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

	dbConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		panic(err)
	}
	dbConfig.ConnConfig.User = c.UserDB.User
	dbConfig.ConnConfig.Password = c.UserDB.Password
	dbConfig.ConnConfig.Host = c.UserDB.Host
	dbConfig.ConnConfig.Port = uint16(c.UserDB.Port)
	dbConfig.ConnConfig.Database = c.UserDB.DBName

	dbPool, err := pgxpool.New(context.Background(), dbConfig.ConnString())
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

	sonyflake.NewSonyflake(sonyflake.Settings{
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
	}
}
