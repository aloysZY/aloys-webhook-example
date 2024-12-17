package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	sugar       *zap.SugaredLogger
	initialized = false
)

// Init 初始化日志记录器
func Init() error {
	if initialized {
		return nil // 如果已经初始化，则不再重复初始化
	}

	// 配置生产环境的日志记录器
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build(
		zap.AddCaller(),                   // 添加调用者信息（文件名和行号）
		zap.AddStacktrace(zap.ErrorLevel), // 在 ErrorLevel 及以上级别添加堆栈跟踪
	)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	sugar = logger.Sugar()
	initialized = true

	// 确保在程序退出时同步日志
	defer func() {
		if initialized {
			logger.Sync()
		}
	}()

	return nil
}

// // Get 获取全局日志记录器
// func Get() *zap.SugaredLogger {
// 	if !initialized {
// 		panic("logger not initialized")
// 	}
// 	return sugar
// }

// WithName 返回带有指定组件名称的日志记录器
func WithName(name string) *zap.SugaredLogger {
	if !initialized {
		panic("logger not initialized")
	}
	return sugar.With(zap.String("component", name))
}
