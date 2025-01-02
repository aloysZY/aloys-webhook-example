package configs

// GlobalConfig 全局配置实例
var globalConfig *Configs

// Configs 定义全局配置
type Configs struct {
	Service     *Service     `yaml:"Service" mapstructure:"Service"`
	Logger      *Logger      `yaml:"Logger" mapstructure:"Logger"`
	Application *Application `yaml:"Application" mapstructure:"Application"`
}

type Logger struct {
	Encoding         string   `yaml:"log_encoding"mapstructure:"log_encoding"` // 日志编码格式（json 或 console）
	LogLevel         string   `yaml:"log_level" mapstructure:"log_level"`
	OutputPaths      []string `yaml:"log_output_paths"mapstructure:"log_output_paths"`             // 日志输出路径，可以是文件、标准输出等
	ErrorOutputPaths []string `yaml:"log_error_output_paths"mapstructure:"log_error_output_paths"` // 错误日志输出路径
	// 日志切割配置
	MaxSize    int `yaml:"max_size" mapstructure:"max_size"`       // 单位 MB，最大文件大小
	MaxBackups int `yaml:"max_backups" mapstructure:"max_backups"` // 保留的最大备份文件数量
	MaxAge     int `yaml:"max_age" mapstructure:"max_age"`         // 保留的最大天数
	// Compress   bool `yaml:"compress" mapstructure:"compress"`       // 是否压缩旧的日志文件
}

type Service struct {
	WebhookBindAddress int    `yaml:"webhook_bind_address" mapstructure:"webhook_bind_address"`
	MetricsBindAddress int    `yaml:"metrics_bind_address" mapstructure:"metrics_bind_address"`
	TLSCertFile        string `yaml:"tls_cert_file"mapstructure:"tls_cert_file"`
	TLSPrivateKeyFile  string `yaml:"tls_private_key_file"mapstructure:"tls_private_key_file"`
	EnablePprof        bool   `yaml:"enable_pprof"mapstructure:"enable_pprof"`
}

type Application struct {
	SidecarImage string `yaml:"sidecar_image"mapstructure:"sidecar_image"`
}

// GetGlobalConfig 提供一个函数来获取全局配置
func GetGlobalConfig() *Configs {
	return globalConfig
}

func InitGlobalConfig(cfg *Configs) {
	globalConfig = cfg
}
