package main

import (
	"errors"
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
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Configs 结构体用于映射配置文件中的键值对
var configPath string // 命令行参数，用于指定配置文件路径

func init() {
	// 初始化 Viper
	// viper.SetConfigFile("internal/configs/config.yaml") // 指定配置文件路径
	viper.SetConfigType("yaml") // 强制使用 YAML 格式
	viper.AutomaticEnv()        // 读取环境变量

	// 添加命令行参数以允许用户指定配置文件路径
	CmdWebhook.Flags().StringVar(&configPath, "config", "", "Path to the configuration file (YAML format).")

	bindFlags(CmdWebhook)
}

func bindFlags(cmd *cobra.Command) {
	// 绑定命令行标志到 Viper
	cmd.Flags().String("tls-cert-file", "", "File containing the default x509 Certificate for HTTPS.")
	_ = viper.BindPFlag("tls_cert_file", cmd.Flags().Lookup("tls-cert-file"))

	cmd.Flags().String("tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	_ = viper.BindPFlag("tls_private_key_file", cmd.Flags().Lookup("tls-private-key-file"))

	cmd.Flags().Int("webhook-bind-address", 0, "Secure port that the webhook-template listens on")
	_ = viper.BindPFlag("webhook_bind_address", cmd.Flags().Lookup("webhook-bind-address"))

	cmd.Flags().Int("metrics-bind-address", 0, "Port that the metrics server listens on.")
	_ = viper.BindPFlag("metrics_bind_address", cmd.Flags().Lookup("metrics-bind-address"))

	cmd.Flags().String("log-level", "", "Set the log level (debug, info, warn, error, dpanic, panic, fatal)")
	_ = viper.BindPFlag("log_level", cmd.Flags().Lookup("log-level"))

	cmd.Flags().Bool("enable-pprof", false, "Enable profiling")
	_ = viper.BindPFlag("enable_pprof", cmd.Flags().Lookup("pprof"))
}

func ladConfig() (*configs.Configs, error) {
	// 如果没有提供配置文件，则使用默认配置
	viper.SetDefault("tls_cert_file", "./certs/tls.crt")
	viper.SetDefault("tls_private_key_file", "./certs/tls.key")
	viper.SetDefault("webhook_bind_address", 9443)
	viper.SetDefault("metrics_bind_address", 8443)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("pprof", false)

	// 根据命令行参数或默认路径设置配置文件
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("internal/configs") // 添加默认路径
		viper.SetConfigName("config")           // 默认配置文件名为 "config.yaml"
	}

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，忽略错误并使用默认值
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			fmt.Println("No configuration file found. Using default values.")
		}
	} else {
		fmt.Printf("Using configuration file: %s\n", viper.ConfigFileUsed())
	}
	cfg := &configs.Configs{}
	// 将配置解码到 Configs 结构体
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return cfg, nil
}

func run(cmd *cobra.Command, args []string) {
	// 加载配置
	cfg, err := ladConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	// 赋值全局配置
	configs.GlobalConfig = cfg
	// 将字符串形式的日志级别转换为 zapcore.Level
	logLevel, err := logger.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid log level specified: %v\n", err)
		os.Exit(1)
	}
	// 初始化日志记录器
	if err := logger.Init(logLevel); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	lg := logger.WithName("main")

	clientSet := util.GetClientSet()
	util.InitializeEventRecorder(clientSet)

	lg.Info("Global variables initialized with values:", // 日志消息
		zap.String("logLevel", cfg.LogLevel),    // 键值对：日志级别
		zap.String("certFile", cfg.CertFile),    // 键值对：证书文件路径
		zap.String("keyFile", cfg.KeyFile),      // 键值对：私钥文件路径
		zap.Int("webhookPort", cfg.WebhookPort), // 键值对：Webhook 端口
		zap.Int("metricsPort", cfg.MetricsPort), // 键值对：Metrics 端口
		zap.Bool("pprof", cfg.EnablePProf),      // 键值对：是否启用 pprof 功能
	)

	// 加载证书
	lg.Debug("Loading TLS certificate and private key files")

	// 启动服务
	metricsServer := api.MetricsStart()
	webhookServer := api.WebhookStart()

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
}

var CmdWebhook = &cobra.Command{
	Use:   "webhook-template",
	Short: "Starts a HTTP server, useful for testing MutatingAdmissionWebhook and ValidatingAdmissionWebhook",
	Long: `Starts a HTTP server, useful for testing MutatingAdmissionWebhook and ValidatingAdmissionWebhook.
After deploying it to Kubernetes cluster, the Administrator needs to create a ValidatingWebhookConfiguration
in the Kubernetes cluster to register remote webhook-template admission controllers.`,
	// Args: cobra.MaximumNArgs(0),
	Run: run,
}

func main() {
	// 解析 goflags 子命令的 flagset，要先解析
	if err := CmdWebhook.Execute(); err != nil {
		// 如果解析 flagset 出错，将 panic 并将 error 信息输出到 lg
		_, _ = fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}

}
