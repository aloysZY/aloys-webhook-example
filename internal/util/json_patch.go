package util

import (
	"github.com/aloys.zy/aloys-webhook-example/internal/global"
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/mattbaird/jsonpatch"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

// GeneratePatchAndResponse 生成 JSON Patch 并返回 AdmissionResponse
func GeneratePatchAndResponse(originalNode, modifiedNode *corev1.Node, allowed bool, warning, message string) *admissionv1.AdmissionResponse {
	// 将原始节点对象序列化为 JSON
	originalNodeBytes, err := json.Marshal(originalNode)
	if err != nil {
		logger.WithName("Json Path").Errorf("failed to marshal originalNode to JSON: %v", err)
		return global.ToV1AdmissionResponse(err)
	}

	// 将修改后的节点对象序列化为 JSON
	modifiedNodeBytes, err := json.Marshal(modifiedNode)
	if err != nil {
		logger.WithName("Json Path").Errorf("failed to marshal modified node to JSON: %v", err)
		return global.ToV1AdmissionResponse(err)
	}

	// 生成从原始节点到修改后节点的 JSON Patch
	patch, err := jsonpatch.CreatePatch(originalNodeBytes, modifiedNodeBytes)
	if err != nil {
		logger.WithName("Json Path").Errorf("failed to create JSON patch: %v", err)
		return global.ToV1AdmissionResponse(err)
	}

	// 将 Patch 序列化为字节数组
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		logger.WithName("Json Path").Errorf("failed to marshal JSON patch: %v", err)
		return global.ToV1AdmissionResponse(err)
	}

	// 构造 AdmissionResponse
	reviewResponse := admissionv1.AdmissionResponse{
		Allowed: allowed,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}

	// 如果有警告信息，添加到 Warnings 中
	if warning != "" {
		reviewResponse.Warnings = []string{warning}
	}

	// 如果有错误信息，添加到 Result.Message 中
	if message != "" {
		reviewResponse.Result = &metav1.Status{
			Message: message,
		}
	}

	return &reviewResponse
}
