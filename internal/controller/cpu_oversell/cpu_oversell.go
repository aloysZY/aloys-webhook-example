package cpu_oversell

import (
	"fmt"
	"math"
	"strconv"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
	"github.com/aloys.zy/aloys-webhook-example/internal/util"
)

const (
	CPUOversell = "cpu_oversell"
)

// MutateCPUOversell 处理节点的 AdmissionReview 请求，根据 cpu_oversell 标签调整 allocatable.cpu
func MutateCPUOversell(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	setupLog := ctrl.Log.WithName("MutateCPUOversell")

	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}

	// 检查请求是否针对节点资源
	if ar.Request.Resource != nodeResource {
		setupLog.Error(nil, "InvalidResource",
			"expected resource to be ", nodeResource,
			"got", ar.Request.Resource)
		return util.GeneratePatchAndResponse(nil, nil, false, "", fmt.Sprintf("expected resource to be %s", nodeResource))
	}

	// 解码传入的节点对象
	var node corev1.Node
	deserializer := setting.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, &node); err != nil {
		setupLog.Error(err, "Failed to decode node object")
		return setting.ToV1AdmissionResponse(err)
	}

	// 保存原始节点对象的副本，用于生成 Patch
	originalNode := node.DeepCopy()

	// 检查是否需要修改 allocatable.cpu
	shouldModify, newAllocatableCPU, err := shouldModifyAllocatableCPU(&node)
	if err != nil {
		// 如果标签无效或解析失败，设置 annotation 为 "false" 并允许请求通过
		updateInvalidLabel(&node, CPUOversell, "false", fmt.Sprintf("Invalid value for %s label on node %s: %v", CPUOversell, node.Name, err))
		return util.GeneratePatchAndResponse(originalNode, &node, true, "", err.Error())
	}

	// 如果不需要修改 allocatable.cpu
	if !shouldModify {
		if shouldUpdateAnnotation(&node, CPUOversell, "false") {
			updateInvalidLabel(&node, CPUOversell, "false", "Added or updated annotation with value 'false'.")
			return util.GeneratePatchAndResponse(originalNode, &node, true, "", "Added or updated annotation with value 'false'.")
		}
		setupLog.V(1).Info("No changes needed for allocatable CPU", "node", node.Name)
		return util.GeneratePatchAndResponse(originalNode, &node, true, "", "No changes needed for allocatable CPU")
	}

	// 将 allocatable.cpu 字符串转换为 milliCPU
	newCPUValue, err := parseCPUStringToMilliCPU(newAllocatableCPU)
	if err != nil {
		setupLog.Error(err, "Error parsing new allocatable CPU value", "node", node.Name)
		return setting.ToV1AdmissionResponse(err)
	}

	// 更新 allocatable.cpu 和注解
	node.Status.Allocatable[corev1.ResourceCPU] = *resource.NewMilliQuantity(newCPUValue, resource.DecimalSI)
	updateInvalidLabel(&node, CPUOversell, "true", fmt.Sprintf("Allocatable CPU updated to %d cores, Annotation %s updated to 'true'.", newCPUValue/1000, CPUOversell))

	// 生成 Patch 并返回，允许请求通过
	return util.GeneratePatchAndResponse(originalNode, &node, true, "", "CPU oversell mutation applied")
}

// shouldModifyAllocatableCPU 检查节点是否有特定的标签，并决定是否修改 allocatable.cpu
func shouldModifyAllocatableCPU(node *corev1.Node) (bool, string, error) {
	setupLog := ctrl.Log.WithName("shouldModifyAllocatableCPU")

	// 检查标签是否存在
	if labels := node.GetLabels(); labels != nil {
		if oversoldCPU, ok := labels[CPUOversell]; ok {
			// 尝试解析标签值为浮点数
			multiplier, err := strconv.ParseFloat(oversoldCPU, 64)
			if err != nil || multiplier <= 0 {
				setupLog.Error(err, "Invalid value for node-oversold-cpu label", "value", oversoldCPU)
				return false, "", err
			}

			// 获取当前的 capacity.cpu 值
			capacityCPU, err := parseCPUQuantity(node.Status.Capacity.Cpu())
			if err != nil {
				setupLog.Error(err, "Error parsing current CPU capacity", "node", node.Name)
				return false, "", err
			}

			// 计算 allocatable.cpu 值
			newAllocatableCPU := float64(capacityCPU.Value()) * multiplier

			// 确保新的 allocatable 不超过合理范围
			if math.IsNaN(newAllocatableCPU) || math.IsInf(newAllocatableCPU, 0) || newAllocatableCPU <= 0 {
				setupLog.V(1).Info("Calculated allocatable CPU is out of reasonable range. Using original allocatable.", "calculated", newAllocatableCPU)
				newAllocatableCPU = float64(capacityCPU.Value())
			}

			// 格式化为字符串
			newCPUFormatted := formatCPUMilliValue(newAllocatableCPU)

			return true, newCPUFormatted, nil
		}
	}

	// 如果标签不存在，返回 false 表示不修改 capacity.cpu
	return false, "", nil
}

// updateInvalidLabel 更新节点的 annotation，并记录事件
func updateInvalidLabel(node *corev1.Node, key, value string, message string) {
	setupLog := ctrl.Log.WithName("updateInvalidLabel")
	util.UpdateAnnotationForInvalidLabel(node, key, value)
	util.EventRecorder().Eventf(node, corev1.EventTypeNormal, "Modified", message)
	setupLog.Info(message, "node", node.Name)
}

// shouldUpdateAnnotation 检查是否需要更新 annotation
func shouldUpdateAnnotation(node *corev1.Node, key, value string) bool {
	if annotations := node.GetAnnotations(); annotations == nil {
		return true
	} else if val, ok := annotations[key]; !ok || val != value {
		return true
	}
	return false
}

// parseCPUQuantity 解析 resource.Quantity 并返回其值
func parseCPUQuantity(cpuQty *resource.Quantity) (*resource.Quantity, error) {
	cpuValue, err := resource.ParseQuantity(cpuQty.String())
	if err != nil {
		return nil, err
	}
	return &cpuValue, nil
}

// formatCPUMilliValue 格式化 float64 表示的 milliCPU 数量为字符串
func formatCPUMilliValue(cpuMilli float64) string {
	return fmt.Sprintf("%fm", cpuMilli)
}

// parseCPUStringToMilliCPU 解析 CPU 字符串并转换为 milliCPU (int64)
func parseCPUStringToMilliCPU(cpuStr string) (int64, error) {
	cpuValue, err := resource.ParseQuantity(cpuStr)
	if err != nil {
		return 0, err
	}
	return cpuValue.MilliValue(), nil
}
