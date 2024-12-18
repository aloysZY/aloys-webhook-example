package configs

import (
	"crypto/tls"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
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
	sugaredLogger := logger.WithName("global.ConfigTLS")

	// Log the paths of the certificate and key files
	sugaredLogger.Debug("Loading TLS certificate and private key from files")
	sugaredLogger.Debugf("CertFile: %s", configs.CertFile)
	sugaredLogger.Debugf("KeyFile: %s", configs.KeyFile)

	// Load the X509 key pair
	sCert, err := tls.LoadX509KeyPair(configs.CertFile, configs.KeyFile)
	if err != nil {
		sugaredLogger.Fatal("Failed to load TLS certificate and private key:", err)
	}
	sugaredLogger.Info("TLS certificate and private key loaded successfully")

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Optionally log the resulting TLS configuration (excluding sensitive information)
	sugaredLogger.Debug("TLS configuration created successfully")
	sugaredLogger.Debugf("Certificates count: %d", len(tlsConfig.Certificates))

	return tlsConfig
}
