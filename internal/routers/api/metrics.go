package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/configs"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsStart(config configs.Configs) *http.Server {
	sugaredLogger := logger.WithName("metrics Start")

	// 启动metrics服务器
	sugaredLogger.Debug("Creating HTTP server mux for metrics endpoints")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("ok"))
		sugaredLogger.Info("Handled /readyz request from", req.RemoteAddr)
	})
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("ok"))
		sugaredLogger.Info("Handled /healthz request from", req.RemoteAddr)
	})

	metricsServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", configs.MetricsPort),
		Handler:        metricsMux,
		TLSConfig:      configs.ConfigTLS(config),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	sugaredLogger.Info("Starting metrics server on port:", configs.MetricsPort)

	go func() {
		defer sugaredLogger.Info("Metrics server goroutine has exited")

		err := metricsServer.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			sugaredLogger.Error("Failed to listen and serve webhook-metrics server:", err)
			panic(err)
		} else if err == http.ErrServerClosed {
			sugaredLogger.Info("Metrics server closed gracefully")
		}
	}()

	return metricsServer
}
