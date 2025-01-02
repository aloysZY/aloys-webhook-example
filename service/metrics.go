package service

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/configs"
	"github.com/aloys.zy/aloys-webhook-example/logger"
	"github.com/aloys.zy/aloys-webhook-example/tls"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func MetricsStart() *http.Server {
	lg := logger.WithName("metrics")

	// 启动metrics服务器
	lg.Debug("Creating HTTP server mux for metrics endpoints")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	// 添加pprof支持
	if configs.GetGlobalConfig().Service.EnablePprof { // 假设有一个配置项来启用或禁用pprof
		metricsMux.Handle("/debug/pprof/", http.DefaultServeMux)
		lg.Info("pprof support enabled on /debug/pprof/")
	}

	// 简化健康检查和就绪检查的处理函数
	handleCheck := func(w http.ResponseWriter, req *http.Request, endpoint string) {
		startTime := time.Now()
		_, _ = w.Write([]byte("ok"))
		// 记录请求完成的日志
		lg.Info(
			"Request completed",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.String("remoteAddr", req.RemoteAddr),
			zap.String("userAgent", req.UserAgent()),
			zap.String("path", req.RequestURI),
			zap.Int("status", http.StatusOK),
			zap.Duration("elapsed_time", time.Since(startTime)),
		)
	}

	metricsMux.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) {
		handleCheck(w, req, "Readyz")
	})
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		handleCheck(w, req, "Healthz")
	})

	metricsServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", configs.GetGlobalConfig().Service.MetricsBindAddress),
		Handler:        metricsMux,
		TLSConfig:      tls.ConfigTLS(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	lg.Info("Starting metrics server on port", zap.Int("port:", configs.GetGlobalConfig().Service.MetricsBindAddress))

	go func() {
		defer lg.Info("Metrics server goroutine has exited")

		err := metricsServer.ListenAndServeTLS("", "")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Error("Failed to listen and serve metrics server:", zap.Error(err))
			panic(err)
		}
	}()

	return metricsServer
}
