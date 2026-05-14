package main

import (
	"flag"
	"fmt"

	"ws-gateway/internal/config"
	"ws-gateway/internal/handler"
	"ws-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/ws-gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer ctx.Hub.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting ws-gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}