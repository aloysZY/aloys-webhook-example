package start

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsStart(configs api.Configs) *http.Server {
	// 启动metrics服务器
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })
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
			logger.WithName("metrics Start.").Error("Failed to listen and serve webhook-metrics server. %v", err)
			panic(err)
		}
	}()
	logger.WithName("metrics Start.").Info("Metrics server started on port：", api.MetricsPort)
	return metricsServer
}
