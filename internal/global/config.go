package global

import (
	"crypto/tls"

	"k8s.io/klog/v2"
)

var (
	CertFile     string
	KeyFile      string
	WebhookPort  int
	SidecarImage string
	MetricsPort  int
)

const (
	TRUE  = "true"
	FALSE = "false"
)

// Configs contains the server (the webhook-template) cert and key.
type Configs struct {
	CertFile string
	KeyFile  string
}

func ConfigTLS(configs Configs) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(configs.CertFile, configs.KeyFile)
	if err != nil {
		klog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}
}

// type PatchOperation struct {
// 	Op    string      `json:"op"`
// 	Path  string      `json:"path"`
// 	Value interface{} `json:"value,omitempty"`
// }

// // GetPatchItem 将json进行模板化
// func GetPatchItem(op string, path string, val interface{}) PatchOperation {
// 	return PatchOperation{
// 		Op:    op,
// 		Path:  path,
// 		Value: val,
// 	}
// }
