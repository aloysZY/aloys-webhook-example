package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aloys.zy/aloys-webhook-example/api"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// MutateNodes 处理节点的 AdmissionReview 请求
func MutateNodes(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(2).Info("mutating nodes...")

	// 定义节点资源
	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}

	if ar.Request.Resource != nodeResource {
		klog.Errorf("expect resource to be %s", nodeResource)
		return nil
	}
	raw := ar.Request.Object.Raw
	node := corev1.Node{}
	deserializer := api.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &node); err != nil {
		klog.Error(err)
		return api.ToV1AdmissionResponse(err)
	}

	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true
	// 检查节点是否具有特定的标签
	// 构建 JSON Patch
	var patches []api.PatchOperation
	if newAllocatableCPU, ok := calculateNewAllocatableCPU(&node); ok {
		patchOps := append(patches,
			api.GetPatchItem("replace", "/status/allocatable/cpu", newAllocatableCPU),
			api.GetPatchItem("replace", "/metadata/annotations/node-oversold-cpu", "true"))
		patchBytes, _ := json.Marshal(patchOps) // json序列号
		reviewResponse.Patch = patchBytes       // 修改的内容添加到path,要json序列号后
		klog.Info("replace allocatable cpu and annotation.")
	} else {
		patchOps := append(patches, api.GetPatchItem("replace", "/metadata/annotations/node-oversold-cpu", "false"))
		patchBytes, _ := json.Marshal(patchOps)
		reviewResponse.Patch = patchBytes
		klog.Info("replace annotation.")
	}
	pt := admissionv1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	return &reviewResponse
}

// calculateNewAllocatableCPU 检查节点是否有特定的标签，并计算新的 allocatable.cpu 值
func calculateNewAllocatableCPU(node *corev1.Node) (string, bool) {
	labels := node.GetLabels()
	if labels == nil {
		return "", false
	}

	oversoldCPU, ok := labels["node-oversold-cpu"]
	if !ok {
		return "", false
	}

	// 解析倍数
	multiplier, err := strconv.Atoi(oversoldCPU)
	if err != nil {
		klog.Errorf("Error parsing multiplier from label: %v", err)
		return "", false
	}

	// 获取当前的 allocatable.cpu 值
	currentCPU, err := strconv.ParseFloat(node.Status.Allocatable.Cpu().String(), 64)
	if err != nil {
		klog.Errorf("Error parsing current CPU value: %v", err)
		return "", false
	}

	// 计算新的 allocatable.cpu 值  * 1000将单位m转化为cpu
	newCPU := currentCPU * float64(multiplier) * 1000

	// 格式化为字符串
	newCPUFormatted := strconv.FormatFloat(newCPU, 'f', -1, 64) + "m"

	return newCPUFormatted, true
}

// ValidatePodContainerLimit 检查 Pod 的 CPU 和内存 limits
func ValidatePodContainerLimit(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(2).Info("validating pods...")

	// 定义 Pod 资源
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	if ar.Request.Resource != podResource {
		klog.Errorf("expect resource to be %s", podResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := scheme.Codecs.UniversalDeserializer()
	_, _, err := deserializer.Decode(raw, nil, &pod)
	if err != nil {
		klog.Errorf("Error decoding pod: %v", err)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to decode pod object",
			},
		}
	}

	// 检查 Pod 的容器
	for i, container := range pod.Spec.Containers {
		if !validateContainerLimits(container) {
			path := fmt.Sprintf("/spec/containers/%d/limits", i)
			klog.Error("Invalid limits: CPU limit must not exceed 6 cores, and memory limit must not exceed 8GiB")
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: fmt.Sprintf("Invalid limits: CPU limit must not exceed 6 cores, and memory limit must not exceed 8GiB"),
					Details: &metav1.StatusDetails{
						Causes: []metav1.StatusCause{
							{
								Field:   path,
								Message: "Invalid limits",
							},
						},
					},
				},
			}
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}

// validateContainerLimits 验证容器的 CPU 和内存 limits 是否超出限制
func validateContainerLimits(container corev1.Container) bool {
	// 检查 CPU limits
	if limit, found := container.Resources.Limits[corev1.ResourceCPU]; found {
		limitStr := limit.String()
		// 判断单位
		if strings.HasSuffix(limitStr, "m") {
			// 如果单位是 "m"，则以 millicores 形式指定
			value, err := strconv.ParseInt(limitStr[:len(limitStr)-1], 10, 64)
			if err != nil {
				// 如果转换失败，返回 false
				return false
			}
			if value > 6000 { // 6000 millicores = 6 cores
				return false
			}
		} else {
			// 如果单位不是 "m"，则以 cores 形式指定
			value, err := strconv.ParseFloat(limitStr, 64)
			if err != nil {
				// 如果转换失败，返回 false
				return false
			}
			if value > 6 { // 6 cores
				return false
			}
		}
	}

	// 检查内存 limits
	if limit, found := container.Resources.Limits[corev1.ResourceMemory]; found {
		if limit.Value() > 8589934592 { // 8589934592 bytes = 8 GiB
			return false
		}
	}

	return true
}
