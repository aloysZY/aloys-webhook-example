package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/configs"
	"github.com/aloys.zy/aloys-webhook-example/internal/metrics"
	"github.com/aloys.zy/aloys-webhook-example/internal/tls"
	ctrl "sigs.k8s.io/controller-runtime"
)

func WebhookStart(cfg *configs.Config) *http.Server {
	setupLog := ctrl.Log.WithName("webhook Start")

	// 创建 HTTP 服务器多路复用器并注册处理函数
	setupLog.V(1).Info("Creating HTTP server mux for webhook endpoints")
	webhook := http.NewServeMux()

	// 注册各个 webhook 处理函数，并包裹上 metrics 中间件
	endpoints := map[string]string{
		"/mutating-cpu-oversell": "ServeMutateCPUOversell",
		"/mutating-pod-dns":      "MutatePodDNSConfig",
		// "/always-allow-delay-5s":    "ServeAlwaysAllowDelayFiveSeconds",
		// "/always-deny":              "ServeAlwaysDeny",
		// "/add-label":                "ServeAddLabel",
		// "/pods":                     "ServePods",
		// "/pods/attach":              "ServeAttachingPods",
		// "/mutating-pods":            "ServeMutatePods",
		// "/mutating-pods-sidecar":    "ServeMutatePodsSidecar",
		// "/configmaps":               "ServeConfigmaps",
		// "/mutating-configmaps":      "ServeMutateConfigmaps",
		// "/custom-resource":          "ServeCustomResource",
		// "/mutating-custom-resource": "ServeMutateCustomResource",
		// "/crd":                      "ServeCRD",
		// "/validating-pod-container-limit": "ServeValidatePodContainerLimit", // Commented out for now
	}

	for endpoint, handlerName := range endpoints {
		handlerFunc := metrics.WithMetrics(getHandlerFuncByName(handlerName))
		webhook.HandleFunc(endpoint, handlerFunc)
		setupLog.Info(
			"Registered webhook endpoint",
			"endpoint", endpoint, // 键值对：endpoint
			"handler", handlerName, // 键值对：handler
		)

	}

	// 创建并配置 HTTP 服务器
	webhookServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.WebhookBindPort),
		TLSConfig:      tls.ConfigTLS(),
		Handler:        webhook,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	setupLog.Info("Starting webhook server on port:", "port:", cfg.WebhookBindPort)

	// 启动服务
	go func() {
		defer setupLog.Info("Webhook server goroutine has exited")
		err := webhookServer.ListenAndServeTLS("", "")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			setupLog.Error(err, "Failed to listen and serve webhook server:")
			panic(err)
		} else if errors.Is(err, http.ErrServerClosed) {
			setupLog.Info("Webhook server closed gracefully")
		}
	}()

	return webhookServer
}
