package server

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/ergoapi/util/exgin"
	"github.com/ergoapi/util/version"
	_ "github.com/ergoapi/util/version/prometheus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/ysicing/default-backend/pkg/templates"
)

func Index(c *gin.Context) {
	host := c.Request.Host
	ip := exgin.RealIP(c)
	c.HTML(403, "ip.html", gin.H{
		"host": host,
		"ip":   ip,
	})
}

func Serve(ctx context.Context) error {
	g := exgin.Init(&exgin.Config{
		Debug:   true,
		Metrics: true,
		Cors:    true,
	})
	g.Use(exgin.ExCors())
	g.Use(exgin.ExLog("/healthz", "/metrics", "/favicon.ico"))
	g.Use(exgin.ExRecovery())
	g.Use(exgin.ExTraceID())
	tpls := template.Must(template.New("").ParseFS(templates.FS, "pages/*.html"))
	g.SetHTMLTemplate(tpls)
	g.GET("/", Index)
	g.GET("/healthz", func(c *gin.Context) {
		exgin.GinsData(c, map[string]string{
			"healthz": "healthz",
		}, nil)
	})
	g.GET("/rv", func(c *gin.Context) {
		v := version.Get()
		exgin.GinsData(c, map[string]string{
			"builddate": v.BuildDate,
			"gitcommit": v.GitCommit,
			"version":   v.GitVersion,
		}, nil)
	})
	g.NoMethod(Index)
	g.NoRoute(Index)
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
			logrus.Errorf("Failed to stop server, error: %s", err)
		}
		logrus.Info("server exited.")
	}()
	logrus.Infof("http listen to %v, pid is %v, version: %v", addr, os.Getpid(), version.GetShortString())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Failed to start http server, error: %s", err)
		return err
	}
	return nil
}
