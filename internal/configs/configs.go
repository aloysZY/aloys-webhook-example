package configs

import (
	"crypto/tls"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"go.uber.org/zap"
)

var GlobalConfig *Configs // 全局配置变量

type Configs struct {
	CertFile     string `mapstructure:"tls_cert_file"`
	KeyFile      string `mapstructure:"tls_private_key_file"`
	LogLevel     string `mapstructure:"log_level"`
	SidecarImage string `mapstructure:"sidecar_image"`
	WebhookPort  int    `mapstructure:"webhook_bind_address"`
	MetricsPort  int    `mapstructure:"metrics_bind_address"`
	EnablePProf  bool   `mapstructure:"pprof"`
}

func ConfigTLS() *tls.Config {
	lg := logger.WithName("global.ConfigTLS")

	// Log the paths of the certificate and key files
	lg.Debug("Loading TLS certificate and private key from files",
		zap.String("CertFile", GlobalConfig.CertFile),
		zap.String("KeyFile", GlobalConfig.KeyFile),
	)

	// Load the X509 key pair
	sCert, err := tls.LoadX509KeyPair(GlobalConfig.CertFile, GlobalConfig.KeyFile)
	if err != nil {
		lg.Fatal("Failed to load TLS certificate and private key:", zap.Error(err))
	}
	lg.Info("TLS certificate and private key loaded successfully")

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Optionally log the resulting TLS configuration (excluding sensitive information)
	lg.Debug("TLS configuration created successfully",
		zap.Int("Certificates count:", len(tlsConfig.Certificates)),
	)

	return tlsConfig
}

// GetGlobalConfig 提供一个函数来获取全局配置
func GetGlobalConfig() *Configs {
	return GlobalConfig
}
