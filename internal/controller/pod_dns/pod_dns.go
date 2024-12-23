package pod_dns

import (
	"fmt"
	"net"
	"reflect"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
	"github.com/aloys.zy/aloys-webhook-example/internal/util"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func MutatePodDNSConfig(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	lg := logger.WithName("MutatePodDNSConfig")

	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	// 检查请求是否针对 pod 资源
	if ar.Request.Resource != podResource {
		lg.Error("InvalidResource",
			zap.Any("expected resource to be ", podResource),
			zap.Any("got", ar.Request.Resource))
		return util.GeneratePatchAndResponse(nil, nil, false, "", fmt.Sprintf("expected resource to be %s", podResource))
	}

	var pod, oldPod corev1.Pod
	deserializer := setting.Codecs.UniversalDeserializer()

	// 解码新对象
	if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, &pod); err != nil {
		lg.Error("Failed to decode new pod object", zap.Error(err))
		return setting.ToV1AdmissionResponse(err)
	}

	// 如果是 UPDATE 操作，解码旧对象
	if ar.Request.Operation == admissionv1.Update {
		if _, _, err := deserializer.Decode(ar.Request.OldObject.Raw, nil, &oldPod); err != nil {
			lg.Error("Failed to decode old pod object", zap.Error(err))
			return setting.ToV1AdmissionResponse(err)
		}

		// 比较 spec 是否相同，如果是 status 更新则忽略
		if reflect.DeepEqual(oldPod.Spec, pod.Spec) {
			util.EventRecorder().Eventf(&pod, corev1.EventTypeNormal, "DeepEqual", "Ignoring status update for pod pod Namespace:%s,pod Name:%s", pod.Namespace, pod.Name)
			lg.Info("Ignoring status update for pod", zap.String("pod Namespace", pod.Namespace), zap.String("pod Name", pod.Name))
			return util.GeneratePatchAndResponse(&pod, nil, true, "", "")
		}
	}
	// pod 就是本次请求的pod，
	originalPod := pod.DeepCopy()

	// 修改 DNS 配置
	if pod.Spec.DNSConfig == nil {
		pod.Spec.DNSConfig = &corev1.PodDNSConfig{}
	}

	// 添加其他 DNS 选项
	pod.Spec.DNSConfig.Options = append(pod.Spec.DNSConfig.Options, []corev1.PodDNSConfigOption{{Name: "timeout", Value: stringPtr("2")}, {Name: "ndots", Value: ptr.To("5")}}...)

	// 配置search
	search := []string{"svc.cluster.local", "cluster.local"}
	namespaceSearch := fmt.Sprintf("%s.svc.cluster.local", pod.Namespace)
	// 如果不存在<namespace>.svc.cluster.local，则插入到第一位
	if !contains(search, namespaceSearch) && pod.Namespace != "" {
		search = append([]string{namespaceSearch}, search...)
	}
	pod.Spec.DNSConfig.Searches = search

	localDnsBindAddress, coreDNSBindAddress, err := util.GetDNSIP()
	if err != nil {
		util.EventRecorder().Eventf(&pod, corev1.EventTypeWarning, "GetDNSIP", "Failed to get DNSIP addresses %v", err)
		lg.Warn("Failed to get DNSIP addresses", zap.Error(err))
		// return setting.ToV1AdmissionResponse(err)
	}
	// 要让kubelet 配置--cluster-dns 才能在pod中是 这个顺序，不然nameserver 10.96.0.10会在上面
	// nameserver 169.254.20.10
	// nameserver 10.96.0.10，
	var nameservers []string
	// 先确认是否为空和是否是一个有效的IP，再添加 localDnsBindAddress 和 coreDNSBindAddress
	if localDnsBindAddress != "" && isValidIP(localDnsBindAddress) {
		nameservers = append(nameservers, localDnsBindAddress)
	}
	if coreDNSBindAddress != "" && isValidIP(coreDNSBindAddress) {
		nameservers = append(nameservers, coreDNSBindAddress)
	}
	// 其实如果要是kubelet 配置--cluster-dns后，肯定会追加到pod.Spec.DNSConfig.Nameservers 配置里面，而且是第一个解析
	pod.Spec.DNSConfig.Nameservers = nameservers

	lg.Info("Mutated DNS configuration for pod",
		zap.String("pod Namespace", pod.Namespace),
		zap.String("pod Name", pod.Name), // 优先使用 pod.Name
		zap.String("pod GenerateName", pod.GenerateName)) // 如果 pod.Name 为空，则可以参考 GenerateName

	// 	根据pod找到对应控制器添加事件信息
	_ = util.GetControllerName(&pod, "Mutated DNS", "Mutated DNS configuration for pod")

	return util.GeneratePatchAndResponse(originalPod, &pod, true, "", "")
}

func stringPtr(s string) *string {
	return &s
}

// isValidIP 检查给定的字符串是否是有效的 IPv4 或 IPv6 地址
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// contains 检查切片中是否包含指定的元素
func contains(slice []string, element string) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}
