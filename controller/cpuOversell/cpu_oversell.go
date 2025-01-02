package cpuOversell

import (
	"fmt"
	"math"
	"strconv"

	"github.com/aloys.zy/aloys-webhook-example/internal/event"
	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
	"github.com/aloys.zy/aloys-webhook-example/logger"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aloys.zy/aloys-webhook-example/internal/util"
)

const (
	CPUOversell = "cpu_oversell"
)

// MutateCPUOversell 处理节点的 AdmissionReview 请求，根据 cpu_oversell 标签调整 allocatable.cpu
func MutateCPUOversell(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	lg := logger.WithName("MutateCPUOversell")

	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}

	// 检查请求是否针对节点资源
	if ar.Request.Resource != nodeResource {
		lg.Error("InvalidResource", zap.Any("expected resource to be ", nodeResource))
		return setting.GeneratePatchAndResponse(nil, nil, false, "", fmt.Sprintf("expected resource to be %s", nodeResource))
	}

	// 解码传入的节点对象,这个每次在请求的都是系统真实的数据，比如我们把allocatable.cpu修改了20，但是这里请求的时候还是4，因为当时分配的系统节点就是4C
	var node corev1.Node
	deserializer := setting.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, &node); err != nil {
		lg.Error("Failed to decode node object", zap.Error(err))
		return setting.ToV1AdmissionResponse(err)
	}

	// 保存原始节点对象的副本，用于生成 Patch
	originalNode := node.DeepCopy()

	// 检查是否需要修改 allocatable.cpu
	shouldModify, newAllocatableCPU, err := shouldModifyAllocatableCPU(&node)
	if err != nil {
		// 如果标签无效或解析失败，设置 annotation 为 "false" 并允许请求通过
		util.UpdateAnnotationForInvalidLabel(&node, CPUOversell, "false")
		event.EventRecorder().Eventf(&node, corev1.EventTypeWarning, "Modified", "Invalid value for %s label on node %s: %v", CPUOversell, node.Name, err)
		lg.Warn("Invalid value for label",
			zap.String(CPUOversell, "false"),
			zap.String("node", node.Name),
			zap.Error(err))
		return setting.GeneratePatchAndResponse(
			originalNode,
			&node,
			true,
			fmt.Sprintf("Warning: Invalid value for %s label on node %s: %v", CPUOversell, node.Name, err),
			err.Error(),
		)
	}

	// 在 shouldModify 为 false 的情况下，实际上并不是因为标签无效（即标签值格式错误或不符合预期），而是因为标签不存在或者根据标签计算后不需要修改 capacity.cpu。因此，而实际上只是不需要进行任何修改
	if !shouldModify {
		// 如果 annotation 不存在或不是 "false"，则设置为 "false"
		if _, ok := node.Annotations[CPUOversell]; !ok || node.Annotations[CPUOversell] != "false" {
			util.UpdateAnnotationForInvalidLabel(&node, CPUOversell, "false")
			event.EventRecorder().Eventf(&node, corev1.EventTypeNormal, "Modified", "Added or updated annotation %s with value 'false', Node Name: %s", CPUOversell, node.Name)
			lg.Info("Added or updated annotation with value 'false'.", zap.String("Node Name", node.Name))
			return setting.GeneratePatchAndResponse(originalNode, &node, true, "", "Added or updated annotation with value 'false'.")
		}
		event.EventRecorder().Eventf(&node, corev1.EventTypeNormal, "NoModification", "No changes needed for allocatable CPU. Node name: %s", node.Name)
		lg.Info("No changes needed for allocatable CPU", zap.String("Node Name", node.Name))
		return setting.GeneratePatchAndResponse(originalNode, &node, true, "", "No changes needed for allocatable CPU")
	}
	// 将 allocatable.cpu 字符串转换为 milliCPU
	newCPUValue, err := parseCPUStringToMilliCPU(newAllocatableCPU)
	if err != nil {
		event.EventRecorder().Eventf(&node, corev1.EventTypeWarning, "ParseError", "Error parsing new allocatable CPU value. Node Name: %s", node.Name)
		lg.Error("Error parsing new allocatable CPU value", zap.Error(err))
		return setting.ToV1AdmissionResponse(err)
	}
	// 更新 allocatable.cpu 和注解
	node.Status.Allocatable[corev1.ResourceCPU] = *resource.NewMilliQuantity(newCPUValue, resource.DecimalSI)
	util.UpdateAnnotationForInvalidLabel(&node, CPUOversell, "true")
	event.EventRecorder().Eventf(&node, corev1.EventTypeNormal, "Modified", "Allocatable CPU updated to %d cores,Annotation %s updated to 'true'.", newCPUValue/1000, CPUOversell)
	lg.Info("Modified allocatable cpu and annotation", zap.Int64("Allocatable_CPU: ", newCPUValue/1000), zap.String(CPUOversell, "true"))

	// 生成 Patch 并返回，允许请求通过
	return setting.GeneratePatchAndResponse(originalNode, &node, true, "", "CPU oversell mutation applied")
}

// shouldModifyAllocatableCPU 检查节点是否有特定的标签，并决定是否修改 allocatable.cpu
func shouldModifyAllocatableCPU(node *corev1.Node) (bool, string, error) {
	lg := logger.WithName("shouldModifyAllocatableCPU")

	// 检查标签是否存在
	if labels := node.GetLabels(); labels != nil {
		if oversoldCPU, ok := labels[CPUOversell]; ok {
			// 尝试解析标签值为浮点数
			multiplier, err := strconv.ParseFloat(oversoldCPU, 64)
			if err != nil || multiplier <= 0 {
				lg.Error("Invalid value for node-oversold-cpu label",
					zap.String("value", oversoldCPU),
					zap.Error(err),
				)
				return false, "", fmt.Errorf("invalid value for node-oversold-cpu label: %v", err)
			}

			// 获取当前的 capacity.cpu 值
			capacityCPU, err := parseCPUQuantity(node.Status.Capacity.Cpu())
			if err != nil {
				lg.Error("Error parsing current CPU capacity", zap.Error(err))
				return false, "", err
			}

			// 计算 allocatable.cpu 值
			newAllocatableCPU := float64(capacityCPU.Value()) * multiplier

			// 确保新的 allocatable 不超过合理范围
			if math.IsNaN(newAllocatableCPU) || math.IsInf(newAllocatableCPU, 0) || newAllocatableCPU <= 0 {
				lg.Warn("Calculated allocatable CPU is out of reasonable range. Using original allocatable.")
				// 如果不合理，就使用capacityCPU的值
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
	return fmt.Sprintf("%f", cpuMilli)
}

// parseCPUStringToMilliCPU 解析 CPU 字符串并转换为 milliCPU (int64)
func parseCPUStringToMilliCPU(cpuStr string) (int64, error) {
	cpuValue, err := resource.ParseQuantity(cpuStr)
	if err != nil {
		return 0, err
	}
	return cpuValue.MilliValue(), nil
}
