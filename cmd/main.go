package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/configs"
	"github.com/aloys.zy/aloys-webhook-example/internal/client"
	"github.com/aloys.zy/aloys-webhook-example/internal/event"
	"github.com/aloys.zy/aloys-webhook-example/logger"
	"github.com/aloys.zy/aloys-webhook-example/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var configPath string // 命令行参数，用于指定配置文件路径

// 初始化 Viper 和 Cobra 命令行参数
func init() {
	// 初始化 Viper
	viper.SetConfigType("yaml") // 强制使用 YAML 格式
	viper.AutomaticEnv()        // 读取环境变量

	// 添加命令行参数以允许用户指定配置文件路径
	CmdWebhook.Flags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file (YAML format).")

	bindFlags(CmdWebhook)

	cfg := &configs.Configs{
		Logger: &configs.Logger{
			Encoding:         "console",          // 设置日志编码格式
			LogLevel:         "debug",            // 设置日志级别
			OutputPaths:      []string{"stdout"}, // 设置日志输出路径
			ErrorOutputPaths: []string{"stderr"}, // 设置错误日志输出路径
		},
	}
	// 初始化日志记录器，默认使用 info 级别
	if err := logger.Init(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.WithName("init").Info("Logger initialized successfully")
}

// 绑定命令行标志到 Viper
func bindFlags(cmd *cobra.Command) {
	viper.SetEnvPrefix("MYAPP")                            // 设置环境变量前缀
	viper.AutomaticEnv()                                   // 启用自动环境变量支持
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 将嵌套路径中的点号替换为下划线

	// 绑定 Webhook 和 Metrics 相关的标志
	cmd.Flags().Int("webhook_bind_address", 9443, "Secure port that the webhook-template listens on")
	viper.BindPFlag("service.webhook_bind_address", cmd.Flags().Lookup("webhook_bind_address"))

	cmd.Flags().Int("metrics_bind_address", 8443, "Port that the metrics server listens on.")
	viper.BindPFlag("service.metrics_bind_address", cmd.Flags().Lookup("metrics_bind_address"))

	// 绑定 TLS 相关的标志
	cmd.Flags().String("tls_cert_file", "./tls/certs/tls.crt", "File containing the default x509 Certificate for HTTPS.")
	viper.BindPFlag("service.tls_cert_file", cmd.Flags().Lookup("tls_cert_file"))

	cmd.Flags().String("tls_private_key_file", "./tls/certs/tls.key", "File containing the default x509 private key matching --tls-cert-file.")
	viper.BindPFlag("service.tls_private_key_file", cmd.Flags().Lookup("tls_private_key_file"))

	// 绑定 Pprof 相关的标志
	cmd.Flags().Bool("enable_pprof", false, "Enable profiling")
	viper.BindPFlag("service.enable_pprof", cmd.Flags().Lookup("enable_pprof"))

	// 绑定 LogConfig 相关的标志，使用嵌套的字段路径
	cmd.Flags().String("log_encoding", "console", "Set the log encoding (json, console)")
	viper.BindPFlag("logger.log_encoding", cmd.Flags().Lookup("log_encoding"))

	cmd.Flags().String("log_level", "info", "Set the log level (debug, info, warn, error, dpanic, panic, fatal)")
	viper.BindPFlag("logger.log_level", cmd.Flags().Lookup("log_level"))

	cmd.Flags().StringSlice("log_output_paths", []string{"stdout"}, "Set the log output paths (e.g., stdout, file:/path/to/logfile)")
	viper.BindPFlag("logger.log_output_paths", cmd.Flags().Lookup("log_output_paths"))

	cmd.Flags().StringSlice("log_error_output_paths", []string{"stderr"}, "Set the log error output paths (e.g., stderr, file:/path/to/errorlogfile)")
	viper.BindPFlag("logger.log_error_output_paths", cmd.Flags().Lookup("log_error_output_paths"))

	cmd.Flags().Int("log_max_size", 100, "Set the maximum size of a log file in MB before it is rolled")
	viper.BindPFlag("logger.log_max_size", cmd.Flags().Lookup("log_max_size"))

	cmd.Flags().Int("log_max_backups", 30, "Set the maximum number of old log files to retain")
	viper.BindPFlag("logger.log_max_backups", cmd.Flags().Lookup("log_max_backups"))

	cmd.Flags().Int("log_max_age", 7, "Set the maximum number of days to retain a log file")
	viper.BindPFlag("logger.log_max_age", cmd.Flags().Lookup("log_max_age"))

	// 绑定 Sidecar Image 相关的标志
	cmd.Flags().String("sidecar_image", "", "Docker image for the sidecar container.")
	viper.BindPFlag("application.sidecar_image", cmd.Flags().Lookup("sidecar_image"))
}

// 启动服务
func startServers(cfg *configs.Configs) (metricsServer, webhookServer *http.Server, err error) {
	metricsServer = service.MetricsStart()
	if metricsServer == nil {
		return nil, nil, fmt.Errorf("failed to start metrics server")
	}

	webhookServer = service.WebhookStart()
	if webhookServer == nil {
		return nil, nil, fmt.Errorf("failed to start webhook server")
	}

	logger.WithName("startServers").Info("Metrics and webhook servers started successfully",
		zap.Int("metricsPort", cfg.Service.MetricsBindAddress),
		zap.Int("webhookPort", cfg.Service.WebhookBindAddress),
	)

	return metricsServer, webhookServer, nil
}

// 处理信号并优雅关闭服务器
func handleSignals(metricsServer, webhookServer *http.Server) {
	componentLogger := logger.WithName("handleSignals")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	componentLogger.Warn("Received shutdown signal, shutting down servers gracefully...")

	// 创建上下文，带有时限
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 webhook 服务器
	if err := webhookServer.Shutdown(ctx); err != nil {
		componentLogger.Error("Error shutting down webhook server", zap.Error(err))
	} else {
		componentLogger.Info("Webhook server shut down successfully")
	}

	// 关闭 metrics 服务器
	if err := metricsServer.Shutdown(ctx); err != nil {
		componentLogger.Error("Error shutting down metrics server", zap.Error(err))
	} else {
		componentLogger.Info("Metrics server shut down successfully")
	}

	// 同步日志
	componentLogger.Sync()
}

// 加载配置
func loadConfig() (*configs.Configs, error) {
	// 设置配置文件路径
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("configs") // 添加默认路径
		viper.SetConfigName("config")  // 默认配置文件名为 "config.yaml"
	}

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			logger.WithName("loadConfig").Info("No configuration file found. Using default values.")
		} else {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
	} else {
		logger.WithName("loadConfig").Info("Using configuration file", zap.String("path", viper.ConfigFileUsed()))
	}

	// 解码配置到 Configs 结构体
	cfg := &configs.Configs{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// 重新配置日志记录器
	if err := logger.Reconfigure(cfg); err != nil {
		return nil, fmt.Errorf("failed to reconfigure logger: %v", err)
	}

	return cfg, nil
}

// 主运行函数
func run(cmd *cobra.Command, args []string) {
	componentLogger := logger.WithName("run")
	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		componentLogger.Error("Failed to load configuration", zap.Error(err))
		os.Exit(1)
	}
	// 赋值全局配置
	configs.InitGlobalConfig(cfg)

	// 初始化 Kubernetes 客户端
	clientSet := client.GetClientSet()
	event.InitializeEventRecorder(clientSet)

	// 启动服务
	metricsServer, webhookServer, err := startServers(cfg)
	if err != nil {
		componentLogger.Error("Failed to start servers", zap.Error(err))
		os.Exit(1)
	}

	// 处理信号并优雅关闭服务器
	handleSignals(metricsServer, webhookServer)
}

var CmdWebhook = &cobra.Command{
	Use:   "webhook-template",
	Short: "Starts a HTTP server, useful for testing MutatingAdmissionWebhook and ValidatingAdmissionWebhook",
	Long: `Starts a HTTP server, useful for testing MutatingAdmissionWebhook and ValidatingAdmissionWebhook.
After deploying it to Kubernetes cluster, the Administrator needs to create a ValidatingWebhookConfiguration
in the Kubernetes cluster to register remote webhook-template admission controllers.`,
	Run: run,
}

func main() {
	if err := CmdWebhook.Execute(); err != nil {
		logger.WithName("main").Error("Error executing command", zap.Error(err))
		os.Exit(1)
	}
}
