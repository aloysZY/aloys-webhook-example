package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsStart(configs api.Configs) *http.Server {
	sugaredLogger := logger.WithName("metrics")
	// 启动metrics服务器
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", api.MetricsPort),
		Handler:        metricsMux,
		TLSConfig:      api.ConfigTLS(configs),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	go func() {
		if err := metricsServer.ListenAndServeTLS("", ""); err != nil {
			sugaredLogger.Error("Failed to listen and serve webhook-metrics server. %v", err)
			panic(err)
		}
		// 调试使用
		// if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// 	klog.Errorf("Failed to listen and serve webhook-metrics server: %v", err)
		// 	os.Exit(1)
		// }
	}()

	sugaredLogger.Info("Metrics server started on port：", api.MetricsPort)
	return metricsServer
}
