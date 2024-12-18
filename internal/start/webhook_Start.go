package start

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/global"
	"github.com/aloys.zy/aloys-webhook-example/internal/handlefunc"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/metrics"
)

func WebhookStart(configs global.Configs) *http.Server {
	webhook := http.NewServeMux()
	webhook.HandleFunc("/mutating-cpu-oversell", metrics.WithMetrics(handlefunc.ServeMutateCPUOversell))
	webhook.HandleFunc("/always-allow-delay-5s", metrics.WithMetrics(handlefunc.ServeAlwaysAllowDelayFiveSeconds))
	webhook.HandleFunc("/always-deny", metrics.WithMetrics(handlefunc.ServeAlwaysDeny))
	webhook.HandleFunc("/add-label", metrics.WithMetrics(handlefunc.ServeAddLabel))
	webhook.HandleFunc("/pods", metrics.WithMetrics(handlefunc.ServePods))
	webhook.HandleFunc("/pods/attach", metrics.WithMetrics(handlefunc.ServeAttachingPods))
	webhook.HandleFunc("/mutating-pods", metrics.WithMetrics(handlefunc.ServeMutatePods))
	webhook.HandleFunc("/mutating-pods-sidecar", metrics.WithMetrics(handlefunc.ServeMutatePodsSidecar))
	webhook.HandleFunc("/configmaps", metrics.WithMetrics(handlefunc.ServeConfigmaps))
	webhook.HandleFunc("/mutating-configmaps", metrics.WithMetrics(handlefunc.ServeMutateConfigmaps))
	webhook.HandleFunc("/custom-resource", metrics.WithMetrics(handlefunc.ServeCustomResource))
	webhook.HandleFunc("/mutating-custom-resource", metrics.WithMetrics(handlefunc.ServeMutateCustomResource))
	webhook.HandleFunc("/crd", metrics.WithMetrics(handlefunc.ServeCRD))
	// webhook.HandleFunc("/validating-pod-container-limit", handlefunc.WithMetrics(handlefunc.ServeValidatePodContainerLimit))

	// 服务启动端口和证书配置
	webhookServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", global.WebhookPort),
		TLSConfig:      global.ConfigTLS(configs),
		Handler:        webhook,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	// 开启go 启动服务
	go func() {
		if err := webhookServer.ListenAndServeTLS("", ""); err != nil {
			logger.WithName("webhook Start").Errorf("Failed to listen and serve webhook server: %v", err)
			panic(err)
		}
	}()
	logger.WithName("webhook Start").Info("Webhook server started on port：", global.WebhookPort)
	return webhookServer
}
