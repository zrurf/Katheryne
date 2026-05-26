package main

import (
	"context"
	"flag"
	"fmt"

	"mem/internal/config"
	"mem/internal/server"
	"mem/internal/svc"
	"mem/mem/mem"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/mem.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// Ensure Qdrant collection exists on startup
	if err := ctx.Qdrant.EnsureCollection(context.Background()); err != nil {
		fmt.Printf("Warning: failed to ensure Qdrant collection: %v\n", err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		mem.RegisterMemServer(grpcServer, server.NewMemServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
