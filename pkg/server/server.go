package server

import (
	"context"
	"fmt"
	"html/template"
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
	"k8s.io/client-go/informers"

	"github.com/ysicing/default-backend/pkg/k8s"
	"github.com/ysicing/default-backend/pkg/templates"
)

func Index(c *gin.Context) {
	host := c.Request.Host
	ip := exgin.RealIP(c)
	s := k8s.StartDeploy(host)
	if s == k8s.NotExistCode {
		c.HTML(200, "notexist.html", gin.H{
			"host": host,
		})
		return
	} else if s == k8s.ExistRunning {
		c.HTML(200, "crash.html", gin.H{
			"host": host,
		})
		return
	} else if s == k8s.ExistStart {
		c.HTML(200, "starting.html", gin.H{
			"host": host,
		})
		return
	}
	exgin.GinsData(c, map[string]interface{}{
		"host": host,
		"ip":   ip,
	}, nil)
}

func Serve(ctx context.Context) error {
	stopChan := make(chan struct{})
	factory := informers.NewSharedInformerFactory(k8s.Client, time.Minute)
	controller := k8s.NewControlller(factory)
	controller.Run(stopChan)
	g := exgin.Init(environ.GetEnv("ENVTYPE", "prod") == "prod")
	g.Use(exgin.ExCors())
	g.Use(exgin.ExLog("/healthz", "/metrics"))
	g.Use(exgin.ExRecovery())
	tpls := template.Must(template.New("").ParseFS(templates.FS, "pages/*.html"))
	g.SetHTMLTemplate(tpls)
	g.GET("/", Index)
	g.GET("/healthz", func(c *gin.Context) {
		exgin.GinsData(c, map[string]string{
			"healthz": "healthz",
		}, nil)
	})
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
		defer close(stopChan)
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			zlog.Error("Failed to stop server, error: %s", err)
		}
		zlog.Info("server exited.")
	}()
	zlog.Info("http listen to %v, pid is %v, version: %v", addr, os.Getpid(), version.GetShortString())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zlog.Error("Failed to start http server, error: %s", err)
		return err
	}

	<-stopChan

	return nil
}
