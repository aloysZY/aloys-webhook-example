package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/handlefunc"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
)

func WebhookStart(configs api.Configs) *http.Server {
	sugaredLogger := logger.WithName("webhook ")
	webhook := http.NewServeMux()
	webhook.HandleFunc("/always-allow-delay-5s", handlefunc.WithMetrics(handlefunc.ServeAlwaysAllowDelayFiveSeconds))
	webhook.HandleFunc("/always-deny", handlefunc.WithMetrics(handlefunc.ServeAlwaysDeny))
	webhook.HandleFunc("/add-label", handlefunc.WithMetrics(handlefunc.ServeAddLabel))
	webhook.HandleFunc("/pods", handlefunc.WithMetrics(handlefunc.ServePods))
	webhook.HandleFunc("/pods/attach", handlefunc.WithMetrics(handlefunc.ServeAttachingPods))
	webhook.HandleFunc("/mutating-pods", handlefunc.WithMetrics(handlefunc.ServeMutatePods))
	webhook.HandleFunc("/mutating-pods-sidecar", handlefunc.WithMetrics(handlefunc.ServeMutatePodsSidecar))
	webhook.HandleFunc("/configmaps", handlefunc.WithMetrics(handlefunc.ServeConfigmaps))
	webhook.HandleFunc("/mutating-configmaps", handlefunc.WithMetrics(handlefunc.ServeMutateConfigmaps))
	webhook.HandleFunc("/custom-resource", handlefunc.WithMetrics(handlefunc.ServeCustomResource))
	webhook.HandleFunc("/mutating-custom-resource", handlefunc.WithMetrics(handlefunc.ServeMutateCustomResource))
	webhook.HandleFunc("/crd", handlefunc.WithMetrics(handlefunc.ServeCRD))
	// webhook.HandleFunc("/validating-pod-container-limit", handlefunc.WithMetrics(handlefunc.ServeValidatePodContainerLimit))
	webhook.HandleFunc("/mutating-node-oversold", handlefunc.WithMetrics(handlefunc.ServeMutateNodeOversold))
	webhook.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })
	webhook.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })

	// 服务启动端口和证书配置
	webhookServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", api.WebhookPort),
		TLSConfig:      api.ConfigTLS(configs),
		Handler:        webhook,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	// 开启go 启动服务
	go func() {
		if err := webhookServer.ListenAndServeTLS("", ""); err != nil {
			sugaredLogger.Errorf("Failed to listen and serve webhook server: %v", err)
			panic(err)
		}
		// 调试
		// if err := webhookServer.ListenAndServe(); err != nil {
		// 	klog.Errorf("Failed to listen and serve webhook server: %v", err)
		// 	panic(err)
		// }
	}()
	sugaredLogger.Info("Webhook server started on port：", api.WebhookPort)
	return webhookServer
}
