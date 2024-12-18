package util

import (
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
	"github.com/mattbaird/jsonpatch"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

// GeneratePatchAndResponse 生成 JSON Patch 并返回 AdmissionResponse
func GeneratePatchAndResponse(originalNode, modifiedNode *corev1.Node, allowed bool, warning, message string) *admissionv1.AdmissionResponse {
	sugaredLogger := logger.WithName("util.GeneratePatchAndResponse")

	// 序列化原始节点对象
	originalNodeBytes, err := marshalNode(originalNode)
	if err != nil {
		sugaredLogger.Errorf("failed to marshal originalNode to JSON: %v", err)
		return setting.ToV1AdmissionResponse(err)
	}

	// 序列化修改后的节点对象
	modifiedNodeBytes, err := marshalNode(modifiedNode)
	if err != nil {
		sugaredLogger.Errorf("failed to marshal modified node to JSON: %v", err)
		return setting.ToV1AdmissionResponse(err)
	}

	// 生成 JSON Patch
	patch, err := createJSONPatch(originalNodeBytes, modifiedNodeBytes)
	if err != nil {
		sugaredLogger.Errorf("failed to create JSON patch: %v", err)
		return setting.ToV1AdmissionResponse(err)
	}

	// 序列化 JSON Patch
	patchBytes, err := marshalPatch(patch)
	if err != nil {
		sugaredLogger.Errorf("failed to marshal JSON patch: %v", err)
		return setting.ToV1AdmissionResponse(err)
	}

	// 构造 AdmissionResponse
	reviewResponse := constructAdmissionResponse(allowed, patchBytes, warning, message)

	return reviewResponse
}

// marshalNode 将 Node 对象序列化为 JSON 字节数组
func marshalNode(node *corev1.Node) ([]byte, error) {
	return json.Marshal(node)
}

// createJSONPatch 生成从原始节点到修改后节点的 JSON Patch
func createJSONPatch(original, modified []byte) ([]jsonpatch.JsonPatchOperation, error) {
	return jsonpatch.CreatePatch(original, modified)
}

// marshalPatch 将 JSON Patch 序列化为字节数组
func marshalPatch(patch []jsonpatch.JsonPatchOperation) ([]byte, error) {
	return json.Marshal(patch)
}

// constructAdmissionResponse 构造并返回 AdmissionResponse
func constructAdmissionResponse(allowed bool, patch []byte, warning, message string) *admissionv1.AdmissionResponse {
	response := admissionv1.AdmissionResponse{
		Allowed: allowed,
		Patch:   patch,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}

	// 如果有警告信息，添加到 Warnings 中
	if warning != "" {
		response.Warnings = append(response.Warnings, warning)
	}

	// 如果有错误信息，添加到 Result.Message 中
	if message != "" {
		response.Result = &metav1.Status{
			Message: message,
		}
	}

	return &response
}
