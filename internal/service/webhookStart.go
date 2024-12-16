package root

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/handlefunc"
	"k8s.io/klog/v2"
)

func WebhookStart(configs api.Configs) *http.Server {

	webhook := http.NewServeMux()
	// 简单的http服务器，可以使用gin
	webhook.HandleFunc("/always-allow-delay-5s", handlefunc.WithMetrics(handlefunc.ServeAlwaysAllowDelayFiveSeconds))
	webhook.HandleFunc("/always-deny", handlefunc.ServeAlwaysDeny)
	webhook.HandleFunc("/add-label", handlefunc.ServeAddLabel)
	webhook.HandleFunc("/pods", handlefunc.ServePods)
	webhook.HandleFunc("/pods/attach", handlefunc.ServeAttachingPods)
	webhook.HandleFunc("/mutating-pods", handlefunc.ServeMutatePods)
	webhook.HandleFunc("/mutating-pods-sidecar", handlefunc.ServeMutatePodsSidecar)
	webhook.HandleFunc("/configmaps", handlefunc.ServeConfigmaps)
	webhook.HandleFunc("/mutating-configmaps", handlefunc.ServeMutateConfigmaps)
	webhook.HandleFunc("/custom-resource", handlefunc.ServeCustomResource)
	webhook.HandleFunc("/mutating-custom-resource", handlefunc.ServeMutateCustomResource)
	webhook.HandleFunc("/crd", handlefunc.ServeCRD)
	webhook.HandleFunc("/validating-pod-container-limit", handlefunc.ServeValidatePodContainerLimit)
	webhook.HandleFunc("/mutating-node-oversold", handlefunc.ServeMutateNodeOversold)
	webhook.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })
	webhook.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })

	// 服务启动端口和证书配置
	webhookServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", api.AppPort),
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
			klog.Errorf("Failed to listen and serve webhook-template server: %v", err)
			panic(err)
		}
	}()
	klog.Info("Webhook server started on port：", api.AppPort)
	return webhookServer
}
