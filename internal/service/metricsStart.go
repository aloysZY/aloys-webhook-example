package service

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

func MetricsStart(configs api.Configs) *http.Server {
	// 启动metrics服务器
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", api.MetricsPort),
		Handler: metricsMux,
		// TLSConfig:      api.ConfigTLS(configs),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("Failed to listen and serve metrics server: %v", err)
			os.Exit(1)
		}
	}()
	klog.Info("Metrics server started on port：", api.MetricsPort)
	return metricsServer
}
