package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/ysicing/default-backend/pkg/server"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		stop()
	}()

	if err := server.Serve(ctx); err != nil {
		logrus.Fatalf("run serve: %v", err)
	}
}
