package api

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/configs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ctrl "sigs.k8s.io/controller-runtime"
)

func MetricsStart(cfg *configs.Config) *http.Server {

	setupLog := ctrl.Log.WithName("metrics Start")

	// 启动metrics服务器
	setupLog.V(1).Info("Creating HTTP server mux for metrics endpoints")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	// 简化健康检查和就绪检查的处理函数
	handleCheck := func(w http.ResponseWriter, req *http.Request, endpoint string) {
		startTime := time.Now()
		_, _ = w.Write([]byte("ok"))
		// 记录请求完成的日志
		setupLog.Info(
			"Request completed",
			"method", req.Method,
			"url", req.URL.String(),
			"remoteAddr", req.RemoteAddr,
			"userAgent", req.UserAgent(),
			"path", req.RequestURI,
			"status", http.StatusOK,
			"elapsed_time", time.Since(startTime))
	}

	metricsMux.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) {
		handleCheck(w, req, "Readyz")
	})
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		handleCheck(w, req, "Healthz")
	})

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsBindPort),
		Handler: metricsMux,
		// TLSConfig:      tls.ConfigTLS(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	setupLog.Info("Starting metrics server on port", "port", cfg.MetricsBindPort)

	go func() {
		defer setupLog.Info("Metrics server goroutine has exited")

		// http监听
		err := metricsServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			setupLog.Error(err, "Failed to listen and serve metrics server:")
			panic(err)
		}
	}()

	return metricsServer
}
