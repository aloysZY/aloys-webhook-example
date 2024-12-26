package pod_dns

import (
	admissionv1 "k8s.io/api/admission/v1"
)

// InitPodDnsConfig 使用init 容器进行注入配置
func InitPodDnsConfig(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	return nil
}
