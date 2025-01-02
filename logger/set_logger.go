package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aloys.zy/aloys-webhook-example/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// getEncoder 根据配置返回合适的编码器
func getEncoder(config zap.Config) zapcore.Encoder {
	switch config.Encoding {
	case "json":
		return zapcore.NewJSONEncoder(config.EncoderConfig)
	default:
		// 启用颜色化输出
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(config.EncoderConfig)
	}
}

// getWriteSyncers 根据配置返回适当的 WriteSyncer
func getWriteSyncers(cfg *configs.Configs) []zapcore.WriteSyncer {
	var writeSyncers []zapcore.WriteSyncer

	if cfg.Logger.Encoding == "console" {
		// 如果是 console 编码，只写入控制台
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	} else {
		// 如果是 json 编码，写入文件
		for _, path := range cfg.Logger.OutputPaths {
			// 提前创建日志文件
			if err := createLogFiles(path); err != nil {
				zap.S().Errorf("failed to create log files: %v", err)
				continue
			}
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				zap.S().Errorf("failed to open log file %s: %v", path, err)
				continue
			}
			writeSyncers = append(writeSyncers, zapcore.AddSync(file))
		}
	}

	return writeSyncers
}

// getErrorWriteSyncers 返回错误日志的 WriteSyncer
func getErrorWriteSyncers(cfg *configs.Configs) []zapcore.WriteSyncer {
	var writeSyncers []zapcore.WriteSyncer

	for _, path := range cfg.Logger.ErrorOutputPaths {
		// 提前创建错误日志文件
		if err := createLogFiles(path); err != nil {
			zap.S().Errorf("failed to create error log files: %v", err)
			continue
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			zap.S().Errorf("failed to open error log file %s: %v", path, err)
			continue
		}
		writeSyncers = append(writeSyncers, zapcore.AddSync(file))
	}

	return writeSyncers
}

// WithName 返回带有指定组件名称的日志记录器，符合 zap.Logger 接口
func WithName(name string) *zap.Logger {
	if !initialized() {
		panic("login not initialized")
	}
	return GetLogger().With(zap.String("component", name))
}

// ParseLogLevel 将字符串形式的日志级别解析为 zapcore.LogLevel
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
	l := lg.Load()
	if l == nil {
		panic("lg not initialized")
	}
	return l.(*zap.Logger)
}

// setLogger 安全地设置全局的 zap.Logger
func setLogger(newLogger *zap.Logger) {
	lg.Store(newLogger)
}

// getOldLogger 安全地获取旧的 zap.Logger
func getOldLogger() *zap.Logger {
	l := lg.Load()
	if l == nil {
		return nil
	}
	return l.(*zap.Logger)
}

// initialized 检查 lg 是否已经初始化
func initialized() bool {
	l := lg.Load()
	return l != nil
}

// createLogFiles 提前创建所有需要的日志文件
func createLogFiles(path string) error {
	// 获取当前工作目录的绝对路径
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	if path != "stdout" && path != "stderr" {
		// 将相对路径转换为绝对路径
		absPath, err := filepath.Abs(filepath.Join(workDir, path))
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path for log file: %v", err)
		}

		// 提取日志文件的目录
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}
	}

	return nil
}
