package server

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/environ"
	"github.com/ergoapi/util/version"
	_ "github.com/ergoapi/util/version/prometheus"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/informers"

	"github.com/ysicing/default-backend/pkg/k8s"
	"github.com/ysicing/default-backend/pkg/templates"
)

func fake(c *gin.Context) bool {
	ua := c.Request.UserAgent()
	ua = strings.ToLower(ua)
	path := c.Request.URL.Path
	if strings.Contains(ua, "python") || strings.Contains(ua, "nss") || strings.Contains(ua, "nmap") || strings.Contains(ua, "censys") {
		return true
	}
	if strings.Contains(ua, "bot") || strings.Contains(ua, "bytespider") || strings.Contains(ua, "android") || strings.Contains(ua, "windows") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "nss") || strings.Contains(ua, "nmap") || strings.Contains(ua, "censys") {
		return true
	}
	if strings.Contains(path, ".php") || strings.Contains(path, "wp") {
		return true
	}
	return false
}

func fakehost(host string) bool {
	defaultHost := environ.GetEnv("DOMAIN_SUFFIX", "ysicing.net")
	return !strings.HasSuffix(host, defaultHost)
}

func Index(c *gin.Context) {
	host := c.Request.Host
	ip := exgin.RealIP(c)
	if len(validation.IsDNS1123Subdomain(host)) != 0 {
		c.HTML(403, "ip.html", gin.H{
			"host": host,
			"ip":   ip,
		})
		return
	}

	s := k8s.StartDeploy(host)
	if fake(c) || fakehost(host) || s == k8s.NotExistCode {
		c.HTML(403, "4xx.html", gin.H{
			"host": host,
			"ip":   ip,
		})
		return
	}

	c.HTML(200, "5xx.html", gin.H{
		"host": host,
		"ip":   ip,
	})
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
	g.NoRoute(Index)
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
