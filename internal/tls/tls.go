package tls

import (
	"crypto/tls"

	ctrl "sigs.k8s.io/controller-runtime"
)

func ConfigTLS() *tls.Config {
	setupLog := ctrl.Log.WithName("config-tls")

	// Log the paths of the certificate and key files
	setupLog.V(1).Info("Loading TLS certificate and private key from files",
		"CertFile", "./certs/tls.crt", "KeyFile", "./certs/tls.key")

	// Load the X509 key pair
	sCert, err := tls.LoadX509KeyPair("./certs/tls.crt", "./certs/tls.key")
	if err != nil {
		setupLog.Error(err, "Failed to load TLS certificate and private key:")
	}
	setupLog.Info("TLS certificate and private key loaded successfully")

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Optionally log the resulting TLS configuration (excluding sensitive information)
	setupLog.V(1).Info("TLS configuration created successfully", "Certificates count:", len(tlsConfig.Certificates))

	return tlsConfig
}
