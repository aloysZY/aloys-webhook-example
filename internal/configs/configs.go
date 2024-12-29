package configs

import (
	"flag"
	"os"
	"sync"

	ubzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Config struct {
	SidecarImage    string
	PprofAddr       string
	EnablePprof     bool
	WebhookBindPort int
	MetricsBindPort int

	// 其他配置项
}

var (
	cfg  *Config
	once sync.Once
)

// InitConfig 初始化配置结构体并解析命令行参数，非基础程序配置添加在这里
func InitConfig() {
	once.Do(func() {
		// 创建一个新的Config实例
		cfg = &Config{}

		// 使用flag包解析命令行参数
		flag.StringVar(&cfg.SidecarImage, "sidecar_image", "", "Docker image for the sidecar container.")
		flag.BoolVar(&cfg.EnablePprof, "enable-pprof", false, "Enable pprof profiling")
		flag.StringVar(&cfg.PprofAddr, "pprof-addr", "localhost:6060", "The address on which to expose the pprof handler")
		flag.IntVar(&cfg.WebhookBindPort, "webhook_bind_address", 9443, "Secure port that the webhook-template listens on")
		flag.IntVar(&cfg.MetricsBindPort, "metrics_bind_address", 8443, "Port that the metrics server listens on.")

		// 定义自定义的 Zap 选项
		opts := zap.Options{
			Development:     false,                                   // 生产环境模式
			Level:           ubzap.NewAtomicLevelAt(ubzap.InfoLevel), // 设置日志级别为 Info
			StacktraceLevel: ubzap.ErrorLevel,                        // 只在 Error 级别及以上添加堆栈跟踪
			Encoder: zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				TimeKey:        "ts",                           // 时间戳字段的键名
				LevelKey:       "level",                        // 日志级别字段的键名
				NameKey:        "logger",                       // 日志记录器名称字段的键名
				CallerKey:      "caller",                       // 调用者信息字段的键名
				MessageKey:     "msg",                          // 日志消息字段的键名
				StacktraceKey:  "stacktrace",                   // 堆栈跟踪字段的键名
				LineEnding:     zapcore.DefaultLineEnding,      // 每行日志的换行符默认是/n
				EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写日志级别
				EncodeTime:     zapcore.RFC3339TimeEncoder,     // 时间戳的编码方式（RFC3339 格式）
				EncodeDuration: zapcore.SecondsDurationEncoder, // 持续时间的编码方式（秒）
				EncodeCaller:   zapcore.ShortCallerEncoder,     // 简短的调用者信息
			}),
			DestWriter: os.Stdout, // 输出到标准输出
			ZapOpts: []ubzap.Option{
				ubzap.AddCaller(), // 添加调用者信息
			},
		}

		// 日志参数绑定到命令行
		opts.BindFlags(flag.CommandLine)

		// 解析命令行参数
		flag.Parse()

		// 应用自定义选项并设置全局日志记录器
		ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	})
}

// GetConfig 返回全局配置实例
func GetConfig() *Config {
	if cfg == nil {
		// 如果cfg未初始化，调用 InitConfig 进行初始化
		InitConfig()
	}
	return cfg
}
