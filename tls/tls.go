package tls

import (
	"crypto/tls"

	"github.com/aloys.zy/aloys-webhook-example/configs"
	"github.com/aloys.zy/aloys-webhook-example/logger"
	"go.uber.org/zap"
)

func ConfigTLS() *tls.Config {
	lg := logger.WithName("global.ConfigTLS")

	// Log the paths of the certificate and key files
	lg.Debug("Loading TLS certificate and private key from files",
		zap.String("CertFile", configs.GetGlobalConfig().Service.TLSCertFile),
		zap.String("KeyFile", configs.GetGlobalConfig().Service.TLSPrivateKeyFile),
	)

	// Load the X509 key pair
	sCert, err := tls.LoadX509KeyPair(configs.GetGlobalConfig().Service.TLSCertFile, configs.GetGlobalConfig().Service.TLSPrivateKeyFile)
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
