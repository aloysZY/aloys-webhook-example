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

	"github.com/aloys.zy/aloys-webhook-example/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/service"
	"github.com/spf13/cobra"
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

		sugaredLogger := logger.WithName("main.CmdWebhook")
		// 加载证书
		configs := api.Configs{
			CertFile: api.CertFile,
			KeyFile:  api.KeyFile,
		}
		// 启动服务
		metricsServer := service.MetricsStart(configs)
		webhookServer := service.WebhookStart(configs)

		// hang
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		// 接收到 os shutdown signal 后，关闭server
		<-signalChan
		sugaredLogger.Infof("Got OS shutdown signal, shutting down webhook-template server gracefully...")
		// _ = server.Shutdown(context.Background())

		// 创建上下文，带有时限
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 关闭webhook服务器
		if err := webhookServer.Shutdown(ctx); err != nil {
			sugaredLogger.Errorf("Error shutting down webhook server: %v", err)
		}
		// 关闭metrics服务器
		if err := metricsServer.Shutdown(ctx); err != nil {
			sugaredLogger.Errorf("Error shutting down metrics server: %v", err)
		}
	},
}

func init() {
	// 定义 goflags 子命令的 flagset，并将 log 作为 goflags 的子命令来使用
	// fs := goflags.NewFlagSet("", goflags.PanicOnError)
	// log.InitFlags() 与 goflags.Parse() 配合使用，可以将 log 作为 goflags 的子命令来使用
	// log.InitFlags(fs)
	// 向 CmdWebhook.Flags() 注入 goflags 子命令的 flagset
	// CmdWebhook.Flags().AddGoFlagSet(fs)

	// 定义 webhook-template 所需要的TLS证书和私钥
	CmdWebhook.Flags().StringVar(&api.CertFile, "tls-cert-file", "/tmp/k8s-webhook-template-server/serving-certs/tls.crt",
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	CmdWebhook.Flags().StringVar(&api.KeyFile, "tls-private-key-file", "/tmp/k8s-webhook-template-server/serving-certs/tls.key",
		"File containing the default x509 private key matching --tls-cert-file.")
	// 定义 webhook-template 所需要的启动端口号，默认为9443，可以由 --port 参数来修改
	CmdWebhook.Flags().IntVar(&api.WebhookPort, "webhook-bind-address", 9443,
		"Secure port that the webhook-template listens on")
	// CmdWebhook.Flags().StringVar(&base.SidecarImage, "sidecar-image", "",
	// 	"Image to be used as the injected sidecar")
	CmdWebhook.Flags().IntVar(&api.MetricsPort, "metrics-bind-address", 8443,
		"WebhookPort that the metrics server listens on.")
}

func main() {
	// 初始化日志记录器
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	sugaredLogger := logger.WithName("main")
	// 解析 goflags 子命令的 flagset
	if err := CmdWebhook.Execute(); err != nil {
		// 如果解析 flagset 出错，将 panic 并将 error 信息输出到 sugaredLogger
		sugaredLogger.Error(err, "Error executing command")
		// sugaredLogger.ErrorS(err, "rootCmd.Execute()")
		os.Exit(1)
	}
}
