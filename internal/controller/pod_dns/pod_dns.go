package pod_dns

import (
	"fmt"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MutatePodDNSConfig(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {

	// 假设此处对 admissionReview 进行路径 DNS 配置检查
	sugaredLogger := logger.WithName("MutatePodDNSConfig")

	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	// 检查请求是否针对pod资源
	if ar.Request.Resource != podResource {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("expected resource to be %s", podResource),
			},
		}
	}

	var pod corev1.Pod
	deserializer := setting.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, &pod); err != nil {
		sugaredLogger.Error(err, "Failed to decode pod object")
		return setting.ToV1AdmissionResponse(err)
	}

	originalPod := pod.DeepCopy()

	if err != nil {
		sugaredLogger.Error(err, "Failed to get clientset")
		return setting.ToV1AdmissionResponse(err)
	}
	_, err = clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
	clientset.CoreV1().Services(corev1).Get()

	xx := &corev1.PodDNSConfig{
		Nameservers: append(pod.Spec.DNSConfig.Nameservers, "xxx"),                                                                // 自定义的 DNS 服务器
		Searches:    pod.Spec.DNSConfig.Searches,                                                                                  // 自定义的搜索域
		Options:     append(pod.Spec.DNSConfig.Options, []corev1.PodDNSConfigOption{{Name: "timeout", Value: stringPtr("2")}}...), //
	}
	sugaredLogger.Infof("Mutated DNS configuration for pod %s/%s", pod.Namespace, pod.Name)

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}

func stringPtr(s string) *string {
	return &s
}
