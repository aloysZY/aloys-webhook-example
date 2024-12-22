package configs

import (
	"crypto/tls"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"go.uber.org/zap"
)

var (
	LogLevel     string
	CertFile     string
	KeyFile      string
	SidecarImage string
	WebhookPort  int
	MetricsPort  int
)

// Configs contains the server (the webhook-template) cert and key.
type Configs struct {
	CertFile string
	KeyFile  string
}

func ConfigTLS(configs Configs) *tls.Config {
	lg := logger.WithName("global.ConfigTLS")

	// Log the paths of the certificate and key files
	lg.Debug("Loading TLS certificate and private key from files",
		zap.String("CertFile", configs.CertFile),
		zap.String("KeyFile", configs.KeyFile),
	)

	// Load the X509 key pair
	sCert, err := tls.LoadX509KeyPair(configs.CertFile, configs.KeyFile)
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
