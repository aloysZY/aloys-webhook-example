package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // 导入 pprof 包，确保 pprof 路由被注册
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/configs"
	"github.com/aloys.zy/aloys-webhook-example/internal/routers/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/util"
	"golang.org/x/net/context"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	// 日志初始化
	setupLog = ctrl.Log.WithName("setup")
)

// 启动服务
func startServers(cfg *configs.Config) (metricsServer, webhookServer *http.Server, err error) {
	webhookServer = api.WebhookStart(cfg)
	if webhookServer == nil {
		return nil, nil, fmt.Errorf("failed to start webhook server")
	}
	metricsServer = api.MetricsStart(cfg)
	if metricsServer == nil {
		return nil, nil, fmt.Errorf("failed to start metrics server")
	}
	// 启用 pprof 服务
	if cfg.EnablePprof {
		go func() {
			log.Printf("Starting pprof server on %s", cfg.PprofAddr)
			err := http.ListenAndServe(fmt.Sprintf("%s", cfg.PprofAddr), nil)
			if err != nil {
				log.Fatalf("Failed to start pprof server: %v", err)
			}
		}()
	}

	setupLog.WithName("startServers").Info("Metrics and webhook servers started successfully",
		"webhookPort", cfg.WebhookBindPort, "metricsPort", cfg.MetricsBindPort)

	return metricsServer, webhookServer, nil
}

// 处理信号并优雅关闭服务器
func handleSignals(metricsServer, webhookServer *http.Server) {
	componentLogger := setupLog.WithName("handleSignals")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	componentLogger.V(1).Info("Received shutdown signal, shutting down servers gracefully...")

	// 创建上下文，带有时限
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 webhook 服务器
	if err := webhookServer.Shutdown(ctx); err != nil {
		componentLogger.Error(err, "Error shutting down webhook server")
	} else {
		componentLogger.Info("Webhook server shut down successfully")
	}

	// 关闭 metrics 服务器
	if err := metricsServer.Shutdown(ctx); err != nil {
		componentLogger.Error(err, "Error shutting down metrics server")
	} else {
		componentLogger.Info("Metrics server shut down successfully")
	}
}

func main() {
	// 初始化配置
	configs.InitConfig()
	cfg := configs.GetConfig()

	// 初始化 Kubernetes 客户端
	err := util.InitClientSet()
	if err != nil {
		setupLog.Error(err, "util.GetClientSet failed")
		os.Exit(1)
	}
	// 初始化event
	util.InitializeEventRecorder()

	// 启动服务
	metricsServer, webhookServer, err := startServers(cfg)
	if err != nil {
		setupLog.Error(err, "Failed to start servers")
		os.Exit(1)
	}

	// 处理信号并优雅关闭服务器
	handleSignals(metricsServer, webhookServer)
}
