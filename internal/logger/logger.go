package logger

import (
	"fmt"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	logger      *zap.Logger
	initialized = false
)

// Init 初始化日志记录器，并设置日志级别
func Init(logLevel zapcore.Level) error {
	if initialized {
		return nil // 如果已经初始化，则不再重复初始化
	}

	// 配置生产环境的日志记录器
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 设置日志级别
	config.Level = zap.NewAtomicLevelAt(logLevel)

	var err error
	logger, err = config.Build(
		zap.AddCaller(),                   // 添加调用者信息（文件名和行号）
		zap.AddStacktrace(zap.ErrorLevel), // 在 ErrorLevel 及以上级别添加堆栈跟踪
	)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	initialized = true

	// 设置 controller-runtime 的全局日志记录器
	log.SetLogger(zapr.NewLogger(logger))

	// 确保在程序退出时同步日志
	defer func() {
		if initialized {
			logger.Sync()
		}
	}()

	return nil
}

// WithName 返回带有指定组件名称的日志记录器，符合 zap.Logger 接口
func WithName(name string) *zap.Logger {
	if !initialized {
		panic("logger not initialized")
	}
	return logger.With(zap.String("component", name))
}

// ParseLogLevel 将字符串形式的日志级别解析为 zapcore.Level
func ParseLogLevel(logLevelStr string) (zapcore.Level, error) {
	switch logLevelStr {
	case "debug":
		return zap.DebugLevel, nil
	case "info":
		return zap.InfoLevel, nil
	case "warn", "warning":
		return zap.WarnLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	case "dpanic":
		return zap.DPanicLevel, nil
	case "panic":
		return zap.PanicLevel, nil
	case "fatal":
		return zap.FatalLevel, nil
	default:
		return zap.InfoLevel, fmt.Errorf("invalid log level: %s", logLevelStr)
	}
}

// GetLogger 返回全局的 zap.Logger
func GetLogger() *zap.Logger {
	if !initialized {
		panic("logger not initialized")
	}
	return logger
}
