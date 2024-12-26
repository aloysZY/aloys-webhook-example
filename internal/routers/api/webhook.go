package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/configs"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/metrics"
	"github.com/aloys.zy/aloys-webhook-example/internal/routers"
	"go.uber.org/zap"
)

func WebhookStart() *http.Server {
	cfg := configs.GetGlobalConfig()
	lg := logger.WithName("webhook Start")

	// 创建 HTTP 服务器多路复用器并注册处理函数
	lg.Debug("Creating HTTP server mux for webhook endpoints")
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
		lg.Info(
			"Registered webhook endpoint",
			zap.String("endpoint", endpoint),   // 键值对：endpoint
			zap.String("handler", handlerName), // 键值对：handler
		)

	}

	// 创建并配置 HTTP 服务器
	webhookServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.WebhookPort),
		TLSConfig:      configs.ConfigTLS(),
		Handler:        webhook,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	lg.Info("Starting webhook server on port:",
		zap.Int("port:", cfg.WebhookPort))

	// 启动服务
	go func() {
		defer lg.Info("Webhook server goroutine has exited")
		err := webhookServer.ListenAndServeTLS("", "")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Error("Failed to listen and serve webhook server:", zap.Error(err))
			panic(err)
		} else if errors.Is(err, http.ErrServerClosed) {
			lg.Info("Webhook server closed gracefully")
		}
	}()

	return webhookServer
}

// 辅助函数：根据名称获取对应的处理函数
func getHandlerFuncByName(name string) http.HandlerFunc {
	switch name {
	case "ServeMutateCPUOversell":
		return routers.ServeMutateCPUOversell
	case "MutatePodDNSConfig":
		return routers.MutatePodDNSConfig
	case "ServeAlwaysAllowDelayFiveSeconds":
		return routers.ServeAlwaysAllowDelayFiveSeconds
	case "ServeAlwaysDeny":
		return routers.ServeAlwaysDeny
	case "ServeAddLabel":
		return routers.ServeAddLabel
	case "ServePods":
		return routers.ServePods
	case "ServeAttachingPods":
		return routers.ServeAttachingPods
	case "ServeMutatePods":
		return routers.ServeMutatePods
	case "ServeMutatePodsSidecar":
		return routers.ServeMutatePodsSidecar
	case "ServeConfigmaps":
		return routers.ServeConfigmaps
	case "ServeMutateConfigmaps":
		return routers.ServeMutateConfigmaps
	case "ServeCustomResource":
		return routers.ServeCustomResource
	case "ServeMutateCustomResource":
		return routers.ServeMutateCustomResource
	case "ServeCRD":
		return routers.ServeCRD
	// case "ServeValidatePodContainerLimit":
	// 	return handlefunc.ServeValidatePodContainerLimit
	default:
		logger.WithName("webhook Start").Warn("Unknown handler name",
			zap.String("name", name))
		return nil
	}
}
