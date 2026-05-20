package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	ctx.BotClient.SetHandler(ctx.MsgHandler)
	ctx.MsgHandler.SetSender(ctx.BotClient)

	if err := ctx.BotClient.Start(); err != nil {
		logx.Errorf("Failed to start bot WS client: %v", err)
		os.Exit(1)
	}

	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx.MsgHandler.StartCleanup(bgCtx, 10*time.Minute)

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

	cancel()
	ctx.BotClient.Stop()
	server.Stop()

	fmt.Println("AI Bot shut down gracefully")
}