package logger

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/configs"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	lg atomic.Value // 使用 atomic.Value 来管理 lg
	mu sync.Mutex   // 用于保护 Reconfigure 操作
)

// Init 初始化日志记录器，并设置日志级别
func Init(cfg *configs.Configs) error {
	if initialized() {
		return nil // 如果已经初始化，则不再重复初始化
	}

	newLogger, err := initNewLogger(cfg)
	if err != nil {
		return err
	}

	setLogger(newLogger)
	log.SetLogger(zapr.NewLogger(newLogger))

	return nil
}

// Reconfigure 重新配置日志记录器
func Reconfigure(cfg *configs.Configs) error {
	mu.Lock()
	defer mu.Unlock()

	// 创建新的日志记录器
	newLogger, err := initNewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to reinitialize lg: %v", err)
	}

	// 关闭旧的日志记录器
	oldLogger := getOldLogger()
	if oldLogger != nil {
		oldLogger.Sync() // 同步旧的日志缓冲区
	}

	// 使用 atomic.Value 替换旧的 lg
	setLogger(newLogger)

	// 等待一段时间，确保所有 goroutine 都已经切换到新的 lg
	time.Sleep(100 * time.Millisecond)

	return nil
}

// initNewLogger 初始化新的日志记录器
func initNewLogger(cfg *configs.Configs) (*zap.Logger, error) {
	// 配置生产环境的日志记录器
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel, err := ParseLogLevel(cfg.Logger.LogLevel)
	if err != nil {
		zap.Error(err)
		return nil, err
	}

	// 设置日志级别
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// 设置编码器
	if cfg.Logger.Encoding == "json" {
		config.Encoding = "json"
	} else {
		config.Encoding = "console"
	}

	// 获取写入器
	writeSyncers := getWriteSyncers(cfg)

	// 创建核心日志配置
	core := zapcore.NewCore(
		getEncoder(config),
		zapcore.NewMultiWriteSyncer(writeSyncers...),
		logLevel,
	)

	// 如果有错误输出路径，创建独立的核心日志配置
	if len(cfg.Logger.ErrorOutputPaths) > 0 && cfg.Logger.Encoding == "json" {
		errorWriteSyncers := getErrorWriteSyncers(cfg)
		if len(errorWriteSyncers) > 0 {
			errorCore := zapcore.NewCore(
				getEncoder(config),
				zapcore.NewMultiWriteSyncer(errorWriteSyncers...),
				zap.ErrorLevel,
			)
			core = zapcore.NewTee(core, errorCore)
		}
	}

	// 创建新的日志记录器
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel), // 为错误日志添加堆栈信息
	}

	newLogger := zap.New(core, options...)

	return newLogger, nil
}
