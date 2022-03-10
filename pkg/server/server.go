package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/environ"
	"github.com/ergoapi/util/version"
	_ "github.com/ergoapi/util/version/prometheus"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Index(c *gin.Context) {
	host := c.Request.Host
	ip := exgin.RealIP(c)
	exgin.GinsData(c, map[string]interface{}{
		"host": host,
		"ip":   ip,
	}, nil)
}

func Serve(ctx context.Context) error {
	g := exgin.Init(environ.GetEnv("ENVTYPE", "prod") == "prod")
	g.Use(exgin.ExCors())
	g.Use(exgin.ExLog())
	g.Use(exgin.ExRecovery())
	g.GET("/", Index)
	g.GET("/metrics", gin.WrapH(promhttp.Handler()))
	g.GET("/rv", func(c *gin.Context) {
		v := version.Get()
		exgin.GinsData(c, map[string]string{
			"builddate": v.BuildDate,
			"gitcommit": v.GitCommit,
			"version":   v.GitVersion,
		}, nil)
	})
	g.NoMethod(func(c *gin.Context) {
		msg := fmt.Sprintf("not found: %v", c.Request.Method)
		exgin.GinsAbortWithCode(c, 404, msg)
	})
	g.NoRoute(func(c *gin.Context) {
		msg := fmt.Sprintf("not found: %v", c.Request.URL.Path)
		exgin.GinsAbortWithCode(c, 404, msg)
	})
	addr := "0.0.0.0:65001"
	srv := &http.Server{
		Addr:    addr,
		Handler: g,
	}
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			zlog.Error("Failed to stop server, error: %s", err)
		}
		zlog.Info("server exited.")
	}()
	zlog.Info("http listen to %v, pid is %v", addr, os.Getpid())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zlog.Error("Failed to start http server, error: %s", err)
		return err
	}
	return nil
}
