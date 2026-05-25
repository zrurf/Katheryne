package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ai-bot/internal/config"
	"ai-bot/internal/handler"
	"ai-bot/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/ai-bot.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(c)

	if err := ctx.Orchestrator.Start(); err != nil {
		logx.Errorf("Failed to start orchestrator: %v", err)
		os.Exit(1)
	}
	defer ctx.Orchestrator.Stop()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("AI Bot starting at %s:%d...\n", c.Host, c.Port)
		server.Start()
	}()

	sig := <-sigCh
	logx.Infof("Received signal %v, shutting down...", sig)

	server.Stop()

	fmt.Println("AI Bot shut down gracefully")
}