package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/ergoapi/zlog"
	"github.com/ysicing/default-backend/pkg/server"
)

func init() {
	cfg := zlog.Config{
		Simple:      true,
		ServiceName: "kubetls",
	}
	zlog.InitZlog(&cfg)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		stop()
	}()

	if err := server.Serve(ctx); err != nil {
		zlog.Fatal("run serve: %v", err)
	}
}
