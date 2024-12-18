package controller

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/aloys.zy/aloys-webhook-example/api"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/util"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CPUOversell = "node-oversold-cpu"

// MutateCPUOversell 处理节点的 AdmissionReview 请求
func MutateCPUOversell(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}

	// 检查请求是否针对节点资源
	if ar.Request.Resource != nodeResource {
		logger.WithName("Mutate Nodes").Errorf("expected resource to be %s", nodeResource)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("expected resource to be %s", nodeResource),
			},
		}
	}

	// 解码传入的节点对象
	raw := ar.Request.Object.Raw
	node := corev1.Node{}
	deserializer := api.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &node); err != nil {
		logger.WithName("Mutate Nodes").Error(err)
		return api.ToV1AdmissionResponse(err)
	}

	// 保存原始节点对象的副本，用于生成 Patch
	originalNode := node.DeepCopy()

	// 调用 shouldModifyAllocatableCPU 函数，检查是否需要修改 allocatable.cpu
	shouldModify, newAllocatableCPU, err := shouldModifyAllocatableCPU(&node)
	if err != nil {
		logger.WithName("Mutate Nodes").Errorf("Error determining if CPU should be modified: %v", err)

		// 如果注解不存在或者不是 "false"，修改为 "false"
		util.UpdateAnnotationForInvalidLabel(&node, CPUOversell, api.FALSE)

		// 生成 Patch 并返回，允许请求通过并包含警告信息
		return util.GeneratePatchAndResponse(
			originalNode,
			&node,
			true,
			fmt.Sprintf("Warning: Invalid value for node-oversold-cpu label on node %s: %v", node.Name, err),
			err.Error(),
		)
	}

	// 如果需要修改 allocatable.cpu
	if shouldModify && newAllocatableCPU != "" {
		cpuValue, err := parseCPUStringToMilliCPU(newAllocatableCPU)
		if err != nil {
			logger.WithName("Mutate Nodes").Errorf("Error parsing new allocatable CPU value: %v", err)
			return api.ToV1AdmissionResponse(err)
		}

		// 更新 allocatable.cpu 和注解
		node.Status.Allocatable[corev1.ResourceCPU] = *resource.NewMilliQuantity(cpuValue, resource.DecimalSI)
		util.UpdateAnnotationForInvalidLabel(&node, CPUOversell, api.TRUE)
		logger.WithName("Mutate Nodes").Info("Modified allocatable cpu and annotation.")
	} else if !shouldModify {
		// 如果注解不存在或者不是 "false"，修改为 "false"
		util.UpdateAnnotationForInvalidLabel(&node, CPUOversell, api.FALSE)
		logger.WithName("Mutate Nodes").Info("Added or updated annotation with value 'false'.")
	}

	// 生成 Patch 并返回，允许请求通过
	return util.GeneratePatchAndResponse(originalNode, &node, true, "", "")
}

// shouldModifyAllocatableCPU 检查节点是否有特定的标签，并决定是否修改 allocatable.cpu
func shouldModifyAllocatableCPU(node *corev1.Node) (bool, string, error) {
	// 检查标签是否存在
	if labels := node.GetLabels(); labels != nil {
		if oversoldCPU, ok := labels["node-oversold-cpu"]; ok {
			// 尝试解析标签值为浮点数
			multiplier, err := strconv.ParseFloat(oversoldCPU, 64)
			if err != nil {
				logger.WithName("Mutate Nodes").Errorf("Invalid value for node-oversold-cpu label: %s", oversoldCPU)
				return false, "", fmt.Errorf("invalid value for node-oversold-cpu label: %v", err)
			}

			// 获取当前的 allocatable.cpu 值
			currentCPU, isMilli, err := parseCPUQuantity(node.Status.Allocatable.Cpu())
			if err != nil {
				logger.WithName("Mutate Nodes").Errorf("Error parsing current CPU value: %v", err)
				return false, "", err
			}

			// 计算新的 allocatable.cpu 值
			newCPU := currentCPU * multiplier

			// 格式化为字符串，根据原始单位决定是否添加 "m"
			newCPUFormatted := formatCPUMilliValue(newCPU, isMilli)

			return true, newCPUFormatted, nil
		}
	}

	// 如果标签不存在，返回 false 表示不修改 allocatable.cpu
	return false, "", nil
}

// parseCPUQuantity 解析 resource.Quantity 并返回 float64 表示的 milliCPU 数量和是否为毫核
func parseCPUQuantity(cpuQty *resource.Quantity) (float64, bool, error) {
	// 将 resource.Quantity 转换为 float64
	cpuMilliValue, err := strconv.ParseFloat(cpuQty.String(), 64)
	if err != nil {
		return 0, false, err
	}

	// 判断是否为毫核单位
	isMilli := strings.HasSuffix(cpuQty.String(), "m")

	return cpuMilliValue, isMilli, nil
}

// formatCPUMilliValue 格式化 float64 表示的 milliCPU 数量为字符串，根据 isMilli 决定是否添加 "m" 单位
func formatCPUMilliValue(cpuMilli float64, isMilli bool) string {
	// 四舍五入到最接近的整数毫核
	roundedMilli := math.Round(cpuMilli)

	// 根据是否为毫核单位选择合适的格式
	if isMilli {
		return fmt.Sprintf("%dm", int64(roundedMilli))
	} else {
		return fmt.Sprintf("%d", int64(roundedMilli))
	}
}

// parseCPUStringToMilliCPU 解析 CPU 字符串并转换为 milliCPU (int64)
func parseCPUStringToMilliCPU(cpuStr string) (int64, error) {
	// 尝试解析为浮点数
	cpuFloat, err := strconv.ParseFloat(cpuStr, 64)
	if err != nil {
		return 0, err
	}

	// 四舍五入并转换为 int64
	cpuInt64 := int64(math.Round(cpuFloat))

	// 检查是否超出 int64 范围
	if cpuInt64 < math.MinInt64 || cpuInt64 > math.MaxInt64 {
		logger.WithName("Mutate Nodes").Errorf("CPU value out of int64 range.")
		return 0, fmt.Errorf("CPU value out of int64 range")
	}

	return cpuInt64, nil
}
