/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/configs"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/routers/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	// "k8s.io/log/v2"
	// TODO: try this library to see if it generates correct json patch
	// https://github.com/mattbaird/jsonpatch
)

// CmdWebhook is used by agnhost Cobra.
var CmdWebhook = &cobra.Command{
	Use:   "webhook-template",
	Short: "Starts a HTTP server, useful for testing MutatingAdmissionWebhook and ValidatingAdmissionWebhook",
	Long: `Starts a HTTP server, useful for testing MutatingAdmissionWebhook and ValidatingAdmissionWebhook.
After deploying it to Kubernetes cluster, the Administrator needs to create a ValidatingWebhookConfiguration
in the Kubernetes cluster to register remote webhook-template admission controllers.`,
	Args: cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		// 将字符串形式的日志级别转换为 zapcore.Level
		logLevel, err := logger.ParseLogLevel(configs.LogLevel)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid log level specified: %v\n", err)
			os.Exit(1)
		}
		// 初始化日志记录器
		if err := logger.Init(logLevel); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			os.Exit(1)
		}
		lg := logger.WithName("global.Init")

		clientSet := util.GetClientSet()
		util.InitializeEventRecorder(clientSet)

		lg.Info(
			"Global variables initialized with values:", // 日志消息
			zap.String("logLevel", configs.LogLevel),    // 键值对：日志级别
			zap.String("certFile", configs.CertFile),    // 键值对：证书文件路径
			zap.String("keyFile", configs.KeyFile),      // 键值对：私钥文件路径
			zap.Int("webhookPort", configs.WebhookPort), // 键值对：Webhook 端口
			zap.Int("metricsPort", configs.MetricsPort), // 键值对：Metrics 端口
		)

		lg = logger.WithName("main")
		// 加载证书
		configs := configs.Configs{
			CertFile: configs.CertFile,
			KeyFile:  configs.KeyFile,
		}
		lg.Debug("Loading TLS certificate and private key files")

		// 启动服务
		metricsServer := api.MetricsStart(configs)
		webhookServer := api.WebhookStart(configs)

		lg.Info("Metrics and webhook servers started successfully")

		// hang
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		// 接收到 os shutdown signal 后，关闭server
		<-signalChan
		lg.Warn("Got OS shutdown signal, shutting down webhook-template server gracefully...")

		// 创建上下文，带有时限
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 关闭webhook服务器
		if err := webhookServer.Shutdown(ctx); err != nil {
			lg.Error("Error shutting down webhook server:", zap.Error(err))
		} else {
			lg.Info("Webhook server shut down successfully")
		}

		// 关闭metrics服务器
		if err := metricsServer.Shutdown(ctx); err != nil {
			lg.Error("Error shutting down metrics server:", zap.Error(err))
		} else {
			lg.Info("Metrics server shut down successfully")
		}
	},
}

func init() {
	// 定义 webhook-template 所需要的TLS证书和私钥
	CmdWebhook.Flags().StringVar(&configs.CertFile, "tls-cert-file", "/certs/tls.crt",
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	CmdWebhook.Flags().StringVar(&configs.KeyFile, "tls-private-key-file", "/certs/tls.key",
		"File containing the default x509 private key matching --tls-cert-file.")
	// 定义 webhook-template 所需要的启动端口号，默认为9443，可以由 --port 参数来修改
	CmdWebhook.Flags().IntVar(&configs.WebhookPort, "webhook-bind-address", 9443,
		"Secure port that the webhook-template listens on")
	CmdWebhook.Flags().IntVar(&configs.MetricsPort, "metrics-bind-address", 8443,
		"WebhookPort that the metrics server listens on.")
	CmdWebhook.Flags().StringVar(&configs.LogLevel, "log-level", "info", "Set the log level (debug, info, warn, error, dpanic, panic, fatal)")
}

func main() {
	// 解析 goflags 子命令的 flagset，要先解析
	if err := CmdWebhook.Execute(); err != nil {
		// 如果解析 flagset 出错，将 panic 并将 error 信息输出到 lg
		_, _ = fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
